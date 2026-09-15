package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Sonny00/terraform-provider-thehive/internal/client"
)

var (
	_ resource.Resource                = &CustomFieldResource{}
	_ resource.ResourceWithImportState = &CustomFieldResource{}
)

type CustomFieldResource struct{ client *client.Client }

type customFieldResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Description types.String `tfsdk:"description"`
	Group       types.String `tfsdk:"group"`
	Type        types.String `tfsdk:"type"`
	Options     types.List   `tfsdk:"options"`
	Mandatory   types.Bool   `tfsdk:"mandatory"`
}

func NewCustomFieldResource() resource.Resource { return &CustomFieldResource{} }

func (r *CustomFieldResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_field"
}

func (r *CustomFieldResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TheHive custom field for cases and alerts.",
		Attributes: map[string]schema.Attribute{
			"id":           schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":         schema.StringAttribute{Description: "Internal name (must be unique).", Required: true},
			"display_name": schema.StringAttribute{Description: "Human-readable label.", Optional: true, Computed: true},
			"description":  schema.StringAttribute{Description: "Description of the field.", Optional: true, Computed: true},
			"group":        schema.StringAttribute{Description: "Field group for UI organization.", Required: true},
			"type": schema.StringAttribute{
				Description:   "Field type: string, integer, boolean, float, or date.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"options": schema.ListAttribute{
				Description: "Predefined values (for string type).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"mandatory": schema.BoolAttribute{
				Description: "Whether the field is mandatory.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (r *CustomFieldResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CustomFieldResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan customFieldResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var opts []string
	if !plan.Options.IsNull() {
		resp.Diagnostics.Append(plan.Options.ElementsAs(ctx, &opts, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	cf, err := r.client.CreateCustomField(ctx, client.CustomFieldInput{
		Name: plan.Name.ValueString(), DisplayName: plan.DisplayName.ValueString(),
		Description: plan.Description.ValueString(), Type: plan.Type.ValueString(),
		Group:   plan.Group.ValueString(),
		Options: opts, Mandatory: plan.Mandatory.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating custom field", err.Error())
		return
	}

	plan.ID = types.StringValue(cf.ID)
	plan.Name = types.StringValue(cf.Name)
	plan.DisplayName = types.StringValue(cf.DisplayName)
	plan.Description = types.StringValue(cf.Description)
	plan.Type = types.StringValue(cf.Type)
	plan.Mandatory = types.BoolValue(cf.Mandatory)
	if len(cf.Options) > 0 {
		ol, d := types.ListValueFrom(ctx, types.StringType, cf.Options)
		resp.Diagnostics.Append(d...)
		plan.Options = ol
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CustomFieldResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state customFieldResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cf, err := r.client.GetCustomField(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading custom field", err.Error())
		return
	}
	state.Name = types.StringValue(cf.Name)
	state.DisplayName = types.StringValue(cf.DisplayName)
	state.Description = types.StringValue(cf.Description)
	state.Type = types.StringValue(cf.Type)
	state.Mandatory = types.BoolValue(cf.Mandatory)
	if len(cf.Options) > 0 {
		ol, d := types.ListValueFrom(ctx, types.StringType, cf.Options)
		resp.Diagnostics.Append(d...)
		state.Options = ol
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *CustomFieldResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan customFieldResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var opts []string
	if !plan.Options.IsNull() {
		resp.Diagnostics.Append(plan.Options.ElementsAs(ctx, &opts, false)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	cf, err := r.client.UpdateCustomField(ctx, plan.ID.ValueString(), client.CustomFieldUpdateInput{
		DisplayName: plan.DisplayName.ValueString(),
		Description: plan.Description.ValueString(),
		Options:     opts,
		Mandatory:   plan.Mandatory.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating custom field", err.Error())
		return
	}
	plan.Name = types.StringValue(cf.Name)
	plan.DisplayName = types.StringValue(cf.DisplayName)
	plan.Description = types.StringValue(cf.Description)
	plan.Mandatory = types.BoolValue(cf.Mandatory)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CustomFieldResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state customFieldResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteCustomField(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting custom field", err.Error())
	}
}

func (r *CustomFieldResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
