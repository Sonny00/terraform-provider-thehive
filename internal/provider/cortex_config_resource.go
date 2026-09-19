package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Sonny00/terraform-provider-thehive/internal/client"
)

var _ resource.Resource = &CortexConfigResource{}

type CortexConfigResource struct{ client *client.Client }

type cortexConfigModel struct {
	ID                  types.String `tfsdk:"id"`
	StatusCheckInterval types.String `tfsdk:"status_check_interval"`
	RefreshDelay        types.String `tfsdk:"refresh_delay"`
	MaxRetryOnError     types.Int64  `tfsdk:"max_retry_on_error"`
	JobTimeout          types.String `tfsdk:"job_timeout"`
	Servers             types.List   `tfsdk:"server"`
}

type cortexServerModel struct {
	Name                  types.String `tfsdk:"name"`
	URL                   types.String `tfsdk:"url"`
	AuthKey               types.String `tfsdk:"auth_key"`
	IncludedOrganisations types.List   `tfsdk:"included_organisations"`
	ExcludedOrganisations types.List   `tfsdk:"excluded_organisations"`
	ConcurrentSubmissions types.Int64  `tfsdk:"concurrent_submissions"`
	MaxResponseSize       types.Int64  `tfsdk:"max_response_size"`
}

var cortexServerAttrTypes = map[string]attr.Type{
	"name":                   types.StringType,
	"url":                    types.StringType,
	"auth_key":               types.StringType,
	"included_organisations": types.ListType{ElemType: types.StringType},
	"excluded_organisations": types.ListType{ElemType: types.StringType},
	"concurrent_submissions": types.Int64Type,
	"max_response_size":      types.Int64Type,
}

func NewCortexConfigResource() resource.Resource { return &CortexConfigResource{} }

func (r *CortexConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cortex_config"
}

func (r *CortexConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the TheHive Cortex connector configuration. This is a singleton resource — only one can exist per instance.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"status_check_interval": schema.StringAttribute{
				Description: "Interval between status checks (e.g. '1 minute').",
				Optional:    true, Computed: true, Default: stringdefault.StaticString("1 minute"),
			},
			"refresh_delay": schema.StringAttribute{
				Description: "Delay between refreshes (e.g. '5 seconds').",
				Optional:    true, Computed: true, Default: stringdefault.StaticString("5 seconds"),
			},
			"max_retry_on_error": schema.Int64Attribute{
				Description: "Maximum retries on error.",
				Optional:    true, Computed: true, Default: int64default.StaticInt64(3),
			},
			"job_timeout": schema.StringAttribute{
				Description: "Maximum time to wait for a job (e.g. '3 hours').",
				Optional:    true, Computed: true, Default: stringdefault.StaticString("3 hours"),
			},
			"server": schema.ListNestedAttribute{
				Description: "Cortex server connections.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":     schema.StringAttribute{Description: "Server name.", Required: true},
						"url":      schema.StringAttribute{Description: "Cortex server URL.", Required: true},
						"auth_key": schema.StringAttribute{Description: "API key for Cortex.", Required: true, Sensitive: true},
						"included_organisations": schema.ListAttribute{
							Description: "TheHive organisations that can use this server. Use [\"*\"] for all.",
							Optional:    true, ElementType: types.StringType,
						},
						"excluded_organisations": schema.ListAttribute{
							Description: "TheHive organisations excluded from this server.",
							Optional:    true, ElementType: types.StringType,
						},
						"concurrent_submissions": schema.Int64Attribute{
							Description: "Max concurrent job submissions.",
							Optional:    true, Computed: true, Default: int64default.StaticInt64(8),
						},
						"max_response_size": schema.Int64Attribute{
							Description: "Max response size in bytes.",
							Optional:    true, Computed: true, Default: int64default.StaticInt64(10485760),
						},
					},
				},
			},
		},
	}
}

func (r *CortexConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CortexConfigResource) toAPIConfig(ctx context.Context, plan cortexConfigModel) (*client.CortexConfig, error) {
	cfg := &client.CortexConfig{
		StatusCheckInterval: plan.StatusCheckInterval.ValueString(),
		RefreshDelay:        plan.RefreshDelay.ValueString(),
		MaxRetryOnError:     int(plan.MaxRetryOnError.ValueInt64()),
		JobTimeout:          plan.JobTimeout.ValueString(),
	}

	if !plan.Servers.IsNull() && !plan.Servers.IsUnknown() {
		var servers []cortexServerModel
		if diags := plan.Servers.ElementsAs(ctx, &servers, false); diags.HasError() {
			return nil, fmt.Errorf("reading servers from plan")
		}
		for _, s := range servers {
			srv := client.CortexServer{
				Name:                         s.Name.ValueString(),
				URL:                          s.URL.ValueString(),
				Auth:                         client.CortexAuth{Type: "bearer", Key: s.AuthKey.ValueString()},
				NumberOfConcurrentSubmission: int(s.ConcurrentSubmissions.ValueInt64()),
				MaxResponseSize:              int(s.MaxResponseSize.ValueInt64()),
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

func (r *CortexConfigResource) toState(ctx context.Context, cfg *client.CortexConfig, state *cortexConfigModel) {
	state.ID = types.StringValue("cortex")
	state.StatusCheckInterval = types.StringValue(cfg.StatusCheckInterval)
	state.RefreshDelay = types.StringValue(cfg.RefreshDelay)
	state.MaxRetryOnError = types.Int64Value(int64(cfg.MaxRetryOnError))
	state.JobTimeout = types.StringValue(cfg.JobTimeout)

	if len(cfg.Servers) == 0 {
		state.Servers = types.ListNull(types.ObjectType{AttrTypes: cortexServerAttrTypes})
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

		obj, _ := types.ObjectValue(cortexServerAttrTypes, map[string]attr.Value{
			"name":                   types.StringValue(s.Name),
			"url":                    types.StringValue(s.URL),
			"auth_key":               types.StringValue(s.Auth.Key),
			"included_organisations": inc,
			"excluded_organisations": exc,
			"concurrent_submissions": types.Int64Value(int64(s.NumberOfConcurrentSubmission)),
			"max_response_size":      types.Int64Value(int64(s.MaxResponseSize)),
		})
		serverVals = append(serverVals, obj)
	}
	state.Servers, _ = types.ListValue(types.ObjectType{AttrTypes: cortexServerAttrTypes}, serverVals)
}

func (r *CortexConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan cortexConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.toAPIConfig(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error building config", err.Error())
		return
	}

	result, err := r.client.UpdateCortexConfig(ctx, *cfg)
	if err != nil {
		resp.Diagnostics.AddError("Error creating cortex config", err.Error())
		return
	}

	r.toState(ctx, result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CortexConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state cortexConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.client.GetCortexConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading cortex config", err.Error())
		return
	}

	r.toState(ctx, cfg, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *CortexConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan cortexConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.toAPIConfig(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error building config", err.Error())
		return
	}

	result, err := r.client.UpdateCortexConfig(ctx, *cfg)
	if err != nil {
		resp.Diagnostics.AddError("Error updating cortex config", err.Error())
		return
	}

	r.toState(ctx, result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CortexConfigResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Singleton: on delete, reset servers to empty but keep global settings.
	_, err := r.client.UpdateCortexConfig(ctx, client.CortexConfig{
		StatusCheckInterval: "1 minute",
		RefreshDelay:        "5 seconds",
		MaxRetryOnError:     3,
		JobTimeout:          "3 hours",
		Servers:             []client.CortexServer{},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error resetting cortex config", err.Error())
	}
}
