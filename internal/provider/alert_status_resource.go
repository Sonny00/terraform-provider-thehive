package provider

import (
	"context"
	"fmt"

	"github.com/Sonny00/terraform-provider-thehive/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &AlertStatusResource{}
	_ resource.ResourceWithImportState = &AlertStatusResource{}
)

type AlertStatusResource struct{ client *client.Client }

type alertStatusResourceModel struct {
	ID    types.String `tfsdk:"id"`
	Value types.String `tfsdk:"value"`
	Stage types.String `tfsdk:"stage"`
	Color types.String `tfsdk:"color"`
}

func NewAlertStatusResource() resource.Resource { return &AlertStatusResource{} }

func (r *AlertStatusResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert_status"
}

func (r *AlertStatusResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a custom alert status in TheHive 5.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "The internal ID of the alert status.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"value": schema.StringAttribute{
				Description:   "The display name of the alert status (e.g. \"TOREVIEW\").",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"stage": schema.StringAttribute{
				Description: "The stage this status belongs to. Determines where the status appears in the alert lifecycle.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("New", "InProgress", "Ignored", "Imported"),
				},
			},
			"color": schema.StringAttribute{
				Description: "The hex color code for this status (e.g. \"#4CAF50\").",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *AlertStatusResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *AlertStatusResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan alertStatusResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := client.AlertStatusInput{Value: plan.Value.ValueString(), Stage: plan.Stage.ValueString()}
	if !plan.Color.IsNull() && !plan.Color.IsUnknown() {
		input.Color = plan.Color.ValueString()
	}
	status, err := r.client.CreateAlertStatus(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Error creating alert status", err.Error())
		return
	}

	plan.ID = types.StringValue(status.ID)
	plan.Value = types.StringValue(status.Value)
	plan.Stage = types.StringValue(status.Stage)
	plan.Color = types.StringValue(status.Color)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *AlertStatusResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state alertStatusResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := r.client.GetAlertStatus(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading alert status", err.Error())
		return
	}
	state.ID = types.StringValue(status.ID)
	state.Value = types.StringValue(status.Value)
	state.Stage = types.StringValue(status.Stage)
	state.Color = types.StringValue(status.Color)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *AlertStatusResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan alertStatusResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state alertStatusResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := client.AlertStatusUpdateInput{}
	if !plan.Color.IsNull() && !plan.Color.IsUnknown() {
		input.Color = plan.Color.ValueString()
	}
	status, err := r.client.UpdateAlertStatus(ctx, state.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Error updating alert status", err.Error())
		return
	}

	plan.ID = types.StringValue(status.ID)
	plan.Value = types.StringValue(status.Value)
	plan.Stage = types.StringValue(status.Stage)
	plan.Color = types.StringValue(status.Color)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *AlertStatusResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state alertStatusResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteAlertStatus(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting alert status", err.Error())
	}
}

func (r *AlertStatusResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
