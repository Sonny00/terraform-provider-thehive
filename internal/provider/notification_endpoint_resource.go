package provider

import (
	"context"
	"fmt"

	"github.com/Sonny00/terraform-provider-thehive/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &notificationEndpointResource{}
	_ resource.ResourceWithImportState = &notificationEndpointResource{}
)

func NewNotificationEndpointResource() resource.Resource {
	return &notificationEndpointResource{}
}

type notificationEndpointResource struct {
	client *client.Client
}

type notificationEndpointResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Type                  types.String `tfsdk:"type"`
	URL                   types.String `tfsdk:"url"`
	Version               types.Int64  `tfsdk:"version"`
	AuthType              types.String `tfsdk:"auth_type"`
	AuthUsername          types.String `tfsdk:"auth_username"`
	AuthPassword          types.String `tfsdk:"auth_password"`
	IncludedOrganisations types.List   `tfsdk:"included_organisations"`
	ExcludedOrganisations types.List   `tfsdk:"excluded_organisations"`
}

func (r *notificationEndpointResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_endpoint"
}

func (r *notificationEndpointResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an organisation-scoped notification endpoint in TheHive 5.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "The endpoint identifier (the endpoint name).",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description:   "The unique name of the endpoint.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"type": schema.StringAttribute{
				Description: "The notification endpoint type.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("webhook", "mattermost", "slack", "teams", "kafka", "redis", "email", "http"),
				},
			},
			"url": schema.StringAttribute{
				Description: "The URL of the endpoint.",
				Required:    true,
			},
			"version": schema.Int64Attribute{
				Description: "The endpoint configuration version.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
			},
			"auth_type": schema.StringAttribute{
				Description: "Authentication type: none, basic, or bearer.",
				Optional:    true,
				Computed:    true,
				Validators:  []validator.String{stringvalidator.OneOf("none", "basic", "bearer")},
			},
			"auth_username": schema.StringAttribute{
				Description: "Username used for basic authentication.",
				Optional:    true,
			},
			"auth_password": schema.StringAttribute{
				Description: "Password or bearer token used for authentication.",
				Optional:    true,
				Sensitive:   true,
			},
			"included_organisations": schema.ListAttribute{
				Description: "TheHive organisations allowed to use this endpoint. An empty list means all organisations.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"excluded_organisations": schema.ListAttribute{
				Description: "TheHive organisations excluded from using this endpoint.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *notificationEndpointResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *notificationEndpointResource) buildInput(ctx context.Context, plan notificationEndpointResourceModel) (client.NotificationEndpointInput, error) {
	input := client.NotificationEndpointInput{
		Name:    plan.Name.ValueString(),
		Type:    plan.Type.ValueString(),
		URL:     plan.URL.ValueString(),
		Version: plan.Version.ValueInt64(),
		Auth: &client.NotificationEndpointAuth{
			Type: plan.AuthType.ValueString(),
		},
		WsConfig: map[string]interface{}{
			"proxy": map[string]interface{}{
				"state": "disabled",
				"auth":  map[string]interface{}{},
			},
			"ssl": map[string]interface{}{
				"loose": map[string]interface{}{
					"acceptAnyCertificate": false,
				},
				"trustManager": map[string]interface{}{},
			},
		},
	}
	if plan.AuthType.IsNull() || plan.AuthType.IsUnknown() {
		input.Auth.Type = "none"
	}
	switch input.Auth.Type {
	case "basic":
		if !plan.AuthUsername.IsNull() && !plan.AuthUsername.IsUnknown() {
			input.Auth.Username = plan.AuthUsername.ValueString()
		}
		if !plan.AuthPassword.IsNull() && !plan.AuthPassword.IsUnknown() {
			input.Auth.Password = plan.AuthPassword.ValueString()
		}
	case "bearer":
		if !plan.AuthPassword.IsNull() && !plan.AuthPassword.IsUnknown() {
			input.Auth.Password = plan.AuthPassword.ValueString()
		}
	}

	var err error
	input.IncludedTheHiveOrganisations, err = endpointListValues(ctx, plan.IncludedOrganisations)
	if err != nil {
		return client.NotificationEndpointInput{}, fmt.Errorf("reading included organisations: %w", err)
	}
	input.ExcludedTheHiveOrganisations, err = endpointListValues(ctx, plan.ExcludedOrganisations)
	if err != nil {
		return client.NotificationEndpointInput{}, fmt.Errorf("reading excluded organisations: %w", err)
	}
	return input, nil
}

func endpointListValues(ctx context.Context, value types.List) ([]string, error) {
	if value.IsNull() || value.IsUnknown() {
		return []string{}, nil
	}
	var result []string
	if diags := value.ElementsAs(ctx, &result, false); diags.HasError() {
		return nil, fmt.Errorf("%s", diags.Errors()[0].Detail())
	}
	return result, nil
}

func endpointListFromStrings(values []string) (types.List, error) {
	elements := make([]attr.Value, len(values))
	for i, value := range values {
		elements[i] = types.StringValue(value)
	}
	result, diags := types.ListValue(types.StringType, elements)
	if diags.HasError() {
		return types.List{}, fmt.Errorf("%s", diags.Errors()[0].Detail())
	}
	return result, nil
}

func (r *notificationEndpointResource) mapToState(result *client.NotificationEndpoint, state *notificationEndpointResourceModel) error {
	state.ID = types.StringValue(result.Name)
	state.Name = types.StringValue(result.Name)
	state.Type = types.StringValue(result.Type)
	state.URL = types.StringValue(result.URL)
	state.Version = types.Int64Value(result.Version)

	if result.Auth != nil {
		state.AuthType = types.StringValue(result.Auth.Type)
		if result.Auth.Username != "" {
			state.AuthUsername = types.StringValue(result.Auth.Username)
		}
	} else {
		state.AuthType = types.StringValue("none")
	}

	var err error
	state.IncludedOrganisations, err = endpointListFromStrings(result.IncludedTheHiveOrganisations)
	if err != nil {
		return fmt.Errorf("mapping included organisations: %w", err)
	}
	state.ExcludedOrganisations, err = endpointListFromStrings(result.ExcludedTheHiveOrganisations)
	if err != nil {
		return fmt.Errorf("mapping excluded organisations: %w", err)
	}
	return nil
}

func (r *notificationEndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan notificationEndpointResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input, err := r.buildInput(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error building notification endpoint input", err.Error())
		return
	}
	if err := r.client.CreateOrUpdateEndpoint(ctx, input); err != nil {
		resp.Diagnostics.AddError("Error creating notification endpoint", err.Error())
		return
	}
	result, err := r.client.GetEndpoint(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading notification endpoint after create", err.Error())
		return
	}
	if err := r.mapToState(result, &plan); err != nil {
		resp.Diagnostics.AddError("Error mapping notification endpoint state", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *notificationEndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state notificationEndpointResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.client.GetEndpoint(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading notification endpoint", err.Error())
		return
	}
	password := state.AuthPassword
	if err := r.mapToState(result, &state); err != nil {
		resp.Diagnostics.AddError("Error mapping notification endpoint state", err.Error())
		return
	}
	state.AuthPassword = password
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *notificationEndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan notificationEndpointResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input, err := r.buildInput(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error building notification endpoint input", err.Error())
		return
	}
	if err := r.client.CreateOrUpdateEndpoint(ctx, input); err != nil {
		resp.Diagnostics.AddError("Error updating notification endpoint", err.Error())
		return
	}
	result, err := r.client.GetEndpoint(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading notification endpoint after update", err.Error())
		return
	}
	if err := r.mapToState(result, &plan); err != nil {
		resp.Diagnostics.AddError("Error mapping notification endpoint state", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *notificationEndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state notificationEndpointResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteEndpoint(ctx, state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting notification endpoint", err.Error())
	}
}

func (r *notificationEndpointResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	result, err := r.client.GetEndpoint(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing notification endpoint", err.Error())
		return
	}
	var state notificationEndpointResourceModel
	if err := r.mapToState(result, &state); err != nil {
		resp.Diagnostics.AddError("Error mapping imported notification endpoint state", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
