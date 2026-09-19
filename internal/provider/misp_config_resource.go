package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Sonny00/terraform-provider-thehive/internal/client"
)

var _ resource.Resource = &MISPConfigResource{}

type MISPConfigResource struct{ client *client.Client }

type mispConfigModel struct {
	ID       types.String `tfsdk:"id"`
	Interval types.String `tfsdk:"interval"`
	Servers  types.List   `tfsdk:"server"`
}

type mispServerModel struct {
	Name                  types.String `tfsdk:"name"`
	URL                   types.String `tfsdk:"url"`
	AuthKey               types.String `tfsdk:"auth_key"`
	Purpose               types.String `tfsdk:"purpose"`
	IncludedOrganisations types.List   `tfsdk:"included_organisations"`
	ExcludedOrganisations types.List   `tfsdk:"excluded_organisations"`
}

var mispServerAttrTypes = map[string]attr.Type{
	"name":                   types.StringType,
	"url":                    types.StringType,
	"auth_key":               types.StringType,
	"purpose":                types.StringType,
	"included_organisations": types.ListType{ElemType: types.StringType},
	"excluded_organisations": types.ListType{ElemType: types.StringType},
}

func NewMISPConfigResource() resource.Resource { return &MISPConfigResource{} }

func (r *MISPConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_misp_config"
}

func (r *MISPConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the TheHive MISP connector configuration. This is a singleton resource — only one can exist per instance.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"interval": schema.StringAttribute{
				Description: "Sync interval (e.g. '1 hour').",
				Optional:    true, Computed: true, Default: stringdefault.StaticString("1 hour"),
			},
			"server": schema.ListNestedAttribute{
				Description: "MISP server connections.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":     schema.StringAttribute{Description: "Server name.", Required: true},
						"url":      schema.StringAttribute{Description: "MISP server URL.", Required: true},
						"auth_key": schema.StringAttribute{Description: "API key for MISP.", Required: true, Sensitive: true},
						"purpose": schema.StringAttribute{
							Description: "Usage purpose: ImportOnly, ExportOnly, or ImportAndExport.",
							Optional:    true, Computed: true, Default: stringdefault.StaticString("ImportAndExport"),
						},
						"included_organisations": schema.ListAttribute{
							Description: "TheHive organisations that can use this server.",
							Optional:    true, ElementType: types.StringType,
						},
						"excluded_organisations": schema.ListAttribute{
							Description: "TheHive organisations excluded from this server.",
							Optional:    true, ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (r *MISPConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected type", fmt.Sprintf("got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *MISPConfigResource) toAPIConfig(ctx context.Context, plan mispConfigModel) (*client.MISPConfig, error) {
	cfg := &client.MISPConfig{
		Interval: plan.Interval.ValueString(),
	}

	if !plan.Servers.IsNull() && !plan.Servers.IsUnknown() {
		var servers []mispServerModel
		if diags := plan.Servers.ElementsAs(ctx, &servers, false); diags.HasError() {
			return nil, fmt.Errorf("reading servers from plan")
		}
		for _, s := range servers {
			srv := client.MISPServer{
				Name:    s.Name.ValueString(),
				URL:     s.URL.ValueString(),
				Auth:    client.MISPAuth{Type: "bearer", Key: s.AuthKey.ValueString()},
				Purpose: s.Purpose.ValueString(),
			}
			if !s.IncludedOrganisations.IsNull() {
				var inc []string
				s.IncludedOrganisations.ElementsAs(ctx, &inc, false)
				srv.IncludedTheHiveOrganisations = inc
			}
			if !s.ExcludedOrganisations.IsNull() {
				var exc []string
				s.ExcludedOrganisations.ElementsAs(ctx, &exc, false)
				srv.ExcludedTheHiveOrganisations = exc
			}
			cfg.Servers = append(cfg.Servers, srv)
		}
	}
	return cfg, nil
}

func (r *MISPConfigResource) toState(ctx context.Context, cfg *client.MISPConfig, state *mispConfigModel) {
	state.ID = types.StringValue("misp")
	state.Interval = types.StringValue(cfg.Interval)

	if len(cfg.Servers) == 0 {
		state.Servers = types.ListNull(types.ObjectType{AttrTypes: mispServerAttrTypes})
		return
	}

	var serverVals []attr.Value
	for _, s := range cfg.Servers {
		inc, _ := types.ListValueFrom(ctx, types.StringType, s.IncludedTheHiveOrganisations)
		exc, _ := types.ListValueFrom(ctx, types.StringType, s.ExcludedTheHiveOrganisations)
		if len(s.IncludedTheHiveOrganisations) == 0 {
			inc = types.ListNull(types.StringType)
		}
		if len(s.ExcludedTheHiveOrganisations) == 0 {
			exc = types.ListNull(types.StringType)
		}

		obj, _ := types.ObjectValue(mispServerAttrTypes, map[string]attr.Value{
			"name":                   types.StringValue(s.Name),
			"url":                    types.StringValue(s.URL),
			"auth_key":               types.StringValue(s.Auth.Key),
			"purpose":                types.StringValue(s.Purpose),
			"included_organisations": inc,
			"excluded_organisations": exc,
		})
		serverVals = append(serverVals, obj)
	}
	state.Servers, _ = types.ListValue(types.ObjectType{AttrTypes: mispServerAttrTypes}, serverVals)
}

func (r *MISPConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan mispConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cfg, err := r.toAPIConfig(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error building config", err.Error())
		return
	}
	result, err := r.client.UpdateMISPConfig(ctx, *cfg)
	if err != nil {
		resp.Diagnostics.AddError("Error creating misp config", err.Error())
		return
	}
	r.toState(ctx, result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *MISPConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state mispConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cfg, err := r.client.GetMISPConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading misp config", err.Error())
		return
	}
	r.toState(ctx, cfg, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *MISPConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan mispConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cfg, err := r.toAPIConfig(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error building config", err.Error())
		return
	}
	result, err := r.client.UpdateMISPConfig(ctx, *cfg)
	if err != nil {
		resp.Diagnostics.AddError("Error updating misp config", err.Error())
		return
	}
	r.toState(ctx, result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *MISPConfigResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	_, err := r.client.UpdateMISPConfig(ctx, client.MISPConfig{
		Interval: "1 hour",
		Servers:  []client.MISPServer{},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error resetting misp config", err.Error())
	}
}
