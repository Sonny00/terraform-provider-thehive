package provider

import (
	"context"
	"fmt"

	"github.com/Sonny00/terraform-provider-thehive/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &functionResource{}
	_ resource.ResourceWithImportState = &functionResource{}
)

func NewFunctionResource() resource.Resource {
	return &functionResource{}
}

type functionResource struct {
	client *client.Client
}

type functionResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Mode       types.String `tfsdk:"mode"`
	Definition types.String `tfsdk:"definition"`
	Config     types.Map    `tfsdk:"config"`
	Types      types.List   `tfsdk:"types"`
}

func (r *functionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_function"
}

func (r *functionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a function in TheHive 5. " +
			"Functions are JavaScript handlers used for notifications, alert feeders, actions, and API extensions.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The internal ID of the function.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the function.",
				Required:    true,
			},
			"mode": schema.StringAttribute{
				Description: "Whether the function is active: Enabled or Disabled.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("Enabled"),
				Validators: []validator.String{
					stringvalidator.OneOf("Enabled", "Disabled"),
				},
			},
			"definition": schema.StringAttribute{
				Description: "The JavaScript code of the function. Must define a handle(input, context) function.",
				Required:    true,
			},
			"config": schema.MapAttribute{
				Description: "Key-value configuration passed to the function at runtime.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     mapdefault.StaticValue(types.MapValueMust(types.StringType, map[string]attr.Value{})),
			},
			"types": schema.ListAttribute{
				Description: "List of function types that define when this function can be used. " +
					"Valid values include: notification, action:case, action:alert, action:task, action:log, api, feeder:alert.",
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *functionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *functionResource) buildInput(ctx context.Context, plan functionResourceModel) (*client.FunctionInput, error) {
	input := client.FunctionInput{
		Name:       plan.Name.ValueString(),
		Mode:       plan.Mode.ValueString(),
		Definition: plan.Definition.ValueString(),
	}

	var functionTypes []string
	diags := plan.Types.ElementsAs(ctx, &functionTypes, false)
	if diags.HasError() {
		return nil, fmt.Errorf("reading function types: %v", diags)
	}
	input.Types = functionTypes

	if !plan.Config.IsNull() && !plan.Config.IsUnknown() {
		config := make(map[string]string)
		diags := plan.Config.ElementsAs(ctx, &config, false)
		if diags.HasError() {
			return nil, fmt.Errorf("reading function config: %v", diags)
		}
		input.Config = config
	}

	return &input, nil
}

func (r *functionResource) mapToState(ctx context.Context, result *client.Function, state *functionResourceModel) error {
	state.ID = types.StringValue(result.ID)
	state.Name = types.StringValue(result.Name)
	state.Mode = types.StringValue(result.Mode)
	state.Definition = types.StringValue(result.Definition)

	typeValues := make([]attr.Value, len(result.Types))
	for i, functionType := range result.Types {
		typeValues[i] = types.StringValue(functionType)
	}
	functionTypes, diags := types.ListValue(types.StringType, typeValues)
	if diags.HasError() {
		return fmt.Errorf("mapping function types to state: %v", diags)
	}
	state.Types = functionTypes

	configValues := make(map[string]attr.Value, len(result.Config))
	for key, value := range result.Config {
		configValues[key] = types.StringValue(value)
	}
	config, diags := types.MapValue(types.StringType, configValues)
	if diags.HasError() {
		return fmt.Errorf("mapping function config to state: %v", diags)
	}
	state.Config = config
	return nil
}

func (r *functionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan functionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, err := r.buildInput(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error building function input", err.Error())
		return
	}

	result, err := r.client.CreateFunction(ctx, *input)
	if err != nil {
		resp.Diagnostics.AddError("Error creating function", err.Error())
		return
	}
	if err := r.mapToState(ctx, result, &plan); err != nil {
		resp.Diagnostics.AddError("Error mapping function state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *functionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state functionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetFunction(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading function", err.Error())
		return
	}
	if err := r.mapToState(ctx, result, &state); err != nil {
		resp.Diagnostics.AddError("Error mapping function state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *functionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan functionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state functionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, err := r.buildInput(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error building function input", err.Error())
		return
	}
	update := client.FunctionUpdateInput{
		Name:       input.Name,
		Mode:       input.Mode,
		Definition: input.Definition,
		Config:     input.Config,
		Types:      input.Types,
	}

	result, err := r.client.UpdateFunction(ctx, state.ID.ValueString(), update)
	if err != nil {
		resp.Diagnostics.AddError("Error updating function", err.Error())
		return
	}
	if err := r.mapToState(ctx, result, &plan); err != nil {
		resp.Diagnostics.AddError("Error mapping function state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *functionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state functionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteFunction(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting function", err.Error())
	}
}

func (r *functionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	result, err := r.client.GetFunction(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing function", err.Error())
		return
	}

	var state functionResourceModel
	if err := r.mapToState(ctx, result, &state); err != nil {
		resp.Diagnostics.AddError("Error mapping imported function state", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
