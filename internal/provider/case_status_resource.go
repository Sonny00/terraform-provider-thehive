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
	_ resource.Resource                = &CaseStatusResource{}
	_ resource.ResourceWithImportState = &CaseStatusResource{}
)

type CaseStatusResource struct{ client *client.Client }

type caseStatusResourceModel struct {
	ID    types.String `tfsdk:"id"`
	Value types.String `tfsdk:"value"`
	Stage types.String `tfsdk:"stage"`
	Color types.String `tfsdk:"color"`
}

func NewCaseStatusResource() resource.Resource { return &CaseStatusResource{} }

func (r *CaseStatusResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_case_status"
}

func (r *CaseStatusResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a custom case status in TheHive 5.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "The internal ID of the case status.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"value": schema.StringAttribute{
				Description:   "The display name of the case status (e.g. \"Pending Analysis\").",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"stage": schema.StringAttribute{
				Description: "The stage this status belongs to. Determines where the status appears in the case lifecycle.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("New", "InProgress", "Indeterminate", "FalsePositive", "TruePositive", "Other", "Duplicated"),
				},
			},
			"color": schema.StringAttribute{
				Description: "The hex color code for this status (e.g. \"#FF5722\").",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *CaseStatusResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CaseStatusResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan caseStatusResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := client.CaseStatusInput{Value: plan.Value.ValueString(), Stage: plan.Stage.ValueString()}
	if !plan.Color.IsNull() && !plan.Color.IsUnknown() {
		input.Color = plan.Color.ValueString()
	}
	status, err := r.client.CreateCaseStatus(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Error creating case status", err.Error())
		return
	}

	plan.ID = types.StringValue(status.ID)
	plan.Value = types.StringValue(status.Value)
	plan.Stage = types.StringValue(status.Stage)
	plan.Color = types.StringValue(status.Color)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CaseStatusResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state caseStatusResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := r.client.GetCaseStatus(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading case status", err.Error())
		return
	}
	state.ID = types.StringValue(status.ID)
	state.Value = types.StringValue(status.Value)
	state.Stage = types.StringValue(status.Stage)
	state.Color = types.StringValue(status.Color)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *CaseStatusResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan caseStatusResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state caseStatusResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := client.CaseStatusUpdateInput{}
	if !plan.Color.IsNull() && !plan.Color.IsUnknown() {
		input.Color = plan.Color.ValueString()
	}
	status, err := r.client.UpdateCaseStatus(ctx, state.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Error updating case status", err.Error())
		return
	}

	plan.ID = types.StringValue(status.ID)
	plan.Value = types.StringValue(status.Value)
	plan.Stage = types.StringValue(status.Stage)
	plan.Color = types.StringValue(status.Color)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CaseStatusResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state caseStatusResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteCaseStatus(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting case status", err.Error())
	}
}

func (r *CaseStatusResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
