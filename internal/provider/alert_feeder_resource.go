package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/Sonny00/terraform-provider-thehive/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &alertFeederResource{}
	_ resource.ResourceWithImportState = &alertFeederResource{}
)

func NewAlertFeederResource() resource.Resource {
	return &alertFeederResource{}
}

type alertFeederResource struct {
	client *client.Client
}

type alertFeederIntervalModel struct {
	Value types.Int64  `tfsdk:"value"`
	Unit  types.String `tfsdk:"unit"`
}

type alertFeederHeaderModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type alertFeederAuthModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Type     types.String `tfsdk:"type"`
}

type alertFeederTimeoutModel struct {
	Value types.Int64  `tfsdk:"value"`
	Unit  types.String `tfsdk:"unit"`
}

type alertFeederResourceModel struct {
	ID              types.String              `tfsdk:"id"`
	Name            types.String              `tfsdk:"name"`
	Description     types.String              `tfsdk:"description"`
	Method          types.String              `tfsdk:"method"`
	URL             types.String              `tfsdk:"url"`
	Interval        *alertFeederIntervalModel `tfsdk:"interval"`
	FunctionName    types.String              `tfsdk:"function_name"`
	Body            types.String              `tfsdk:"body"`
	Headers         []alertFeederHeaderModel  `tfsdk:"headers"`
	Enabled         types.Bool                `tfsdk:"enabled"`
	Auth            *alertFeederAuthModel     `tfsdk:"auth"`
	RequestTimeout  *alertFeederTimeoutModel  `tfsdk:"request_timeout"`
	ResponseMaxSize types.Int64               `tfsdk:"response_max_size"`
}

func (r *alertFeederResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert_feeder"
}

func (r *alertFeederResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an alert feeder that periodically fetches an external HTTP API and creates alerts using a TheHive function.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "The alert feeder name, used as its identifier.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name":        schema.StringAttribute{Description: "The alert feeder name.", Required: true},
			"description": schema.StringAttribute{Description: "A description of the feeder.", Optional: true},
			"method":      schema.StringAttribute{Description: "HTTP method used to fetch data.", Required: true},
			"url":         schema.StringAttribute{Description: "URL to fetch data from.", Required: true},
			"function_name": schema.StringAttribute{
				Description: "Name of the TheHive function that processes fetched data into alerts.",
				Required:    true,
			},
			"interval": schema.SingleNestedAttribute{
				Description: "How often the feeder fetches data.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"value": schema.Int64Attribute{Description: "Interval quantity.", Required: true},
					"unit":  schema.StringAttribute{Description: "Interval unit.", Required: true},
				},
			},
			"body": schema.StringAttribute{
				Description: "Optional JSON object sent as the request body.",
				Optional:    true,
			},
			"headers": schema.ListNestedAttribute{
				Description: "Custom HTTP headers to include in the request.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key":   schema.StringAttribute{Description: "Header name.", Required: true},
						"value": schema.StringAttribute{Description: "Header value.", Required: true, Sensitive: true},
					},
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the feeder is active.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"auth": schema.SingleNestedAttribute{
				Description: "Optional authentication configuration.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"type":     schema.StringAttribute{Description: "Authentication type.", Required: true},
					"username": schema.StringAttribute{Description: "Username for authentication.", Optional: true},
					"password": schema.StringAttribute{Description: "Password or token.", Optional: true, Sensitive: true},
				},
			},
			"request_timeout": schema.SingleNestedAttribute{
				Description: "Optional timeout for each request.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"value": schema.Int64Attribute{Description: "Timeout quantity.", Required: true},
					"unit":  schema.StringAttribute{Description: "Timeout unit.", Required: true},
				},
			},
			"response_max_size": schema.Int64Attribute{
				Description: "Maximum response size in bytes.",
				Optional:    true,
			},
		},
	}
}

func (r *alertFeederResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *alertFeederResource) buildInput(plan alertFeederResourceModel) (client.AlertFeederInput, error) {
	if plan.Interval == nil {
		return client.AlertFeederInput{}, fmt.Errorf("interval is required")
	}

	input := client.AlertFeederInput{
		Name:         plan.Name.ValueString(),
		Description:  plan.Description.ValueString(),
		Method:       plan.Method.ValueString(),
		URL:          plan.URL.ValueString(),
		FunctionName: plan.FunctionName.ValueString(),
		Enabled:      plan.Enabled.ValueBool(),
		Interval: client.AlertFeederInterval{
			Value: plan.Interval.Value.ValueInt64(),
			Unit:  plan.Interval.Unit.ValueString(),
		},
	}

	if !plan.Body.IsNull() && !plan.Body.IsUnknown() {
		var body map[string]interface{}
		if err := json.Unmarshal([]byte(plan.Body.ValueString()), &body); err != nil {
			return client.AlertFeederInput{}, fmt.Errorf("decoding body JSON: %w", err)
		}
		input.Body = body
	}
	for _, header := range plan.Headers {
		input.Headers = append(input.Headers, client.AlertFeederHeader{
			Key:   header.Key.ValueString(),
			Value: header.Value.ValueString(),
		})
	}
	if plan.Auth != nil {
		input.Auth = &client.AlertFeederAuth{
			Type:     plan.Auth.Type.ValueString(),
			Username: plan.Auth.Username.ValueString(),
			Password: plan.Auth.Password.ValueString(),
		}
	}
	if plan.RequestTimeout != nil {
		input.RequestTimeout = &client.AlertFeederTimeout{
			Value: plan.RequestTimeout.Value.ValueInt64(),
			Unit:  plan.RequestTimeout.Unit.ValueString(),
		}
	}
	if !plan.ResponseMaxSize.IsNull() && !plan.ResponseMaxSize.IsUnknown() {
		value := plan.ResponseMaxSize.ValueInt64()
		input.ResponseMaxSize = &value
	}
	return input, nil
}

func (r *alertFeederResource) mapToState(result *client.AlertFeeder, state *alertFeederResourceModel) error {
	if result.Name == "" {
		return fmt.Errorf("API response did not include a name")
	}
	state.ID = types.StringValue(result.Name)
	state.Name = types.StringValue(result.Name)
	if !state.Description.IsNull() && !state.Description.IsUnknown() && result.Description != "" {
		state.Description = types.StringValue(result.Description)
	}
	if result.Method != "" {
		state.Method = types.StringValue(result.Method)
	}
	if result.URL != "" {
		state.URL = types.StringValue(result.URL)
	}
	if result.Function.Name != "" {
		state.FunctionName = types.StringValue(result.Function.Name)
	}
	state.Enabled = types.BoolValue(result.Enabled)
	if state.Interval == nil {
		state.Interval = &alertFeederIntervalModel{
			Value: types.Int64Value(result.Interval.Value),
			Unit:  types.StringValue(result.Interval.Unit),
		}
	}

	if !state.Body.IsNull() && !state.Body.IsUnknown() && result.Body != nil {
		body, err := json.Marshal(result.Body)
		if err != nil {
			return fmt.Errorf("encoding response body: %w", err)
		}
		var configuredBody map[string]interface{}
		if json.Unmarshal([]byte(state.Body.ValueString()), &configuredBody) != nil ||
			!reflect.DeepEqual(configuredBody, result.Body) {
			state.Body = types.StringValue(string(body))
		}
	}

	if state.Headers != nil && result.Headers != nil {
		state.Headers = make([]alertFeederHeaderModel, len(result.Headers))
		for i, header := range result.Headers {
			state.Headers[i] = alertFeederHeaderModel{
				Key:   types.StringValue(header.Key),
				Value: types.StringValue(header.Value),
			}
		}
	}
	if result.Auth != nil && result.Auth.Type != "none" && state.Auth != nil && !state.Auth.Type.IsNull() && !state.Auth.Type.IsUnknown() {
		if result.Auth.Type != "" {
			state.Auth.Type = types.StringValue(result.Auth.Type)
		}
		if !state.Auth.Username.IsNull() && result.Auth.Username != "" {
			state.Auth.Username = types.StringValue(result.Auth.Username)
		}
		if !state.Auth.Password.IsNull() && result.Auth.Password != "" {
			state.Auth.Password = types.StringValue(result.Auth.Password)
		}
	}
	if result.RequestTimeout != nil && state.RequestTimeout != nil {
		state.RequestTimeout = &alertFeederTimeoutModel{
			Value: types.Int64Value(result.RequestTimeout.Value),
			Unit:  types.StringValue(result.RequestTimeout.Unit),
		}
	}
	if result.ResponseMaxSize != nil && !state.ResponseMaxSize.IsNull() && !state.ResponseMaxSize.IsUnknown() {
		state.ResponseMaxSize = types.Int64Value(*result.ResponseMaxSize)
	}
	return nil
}

func (r *alertFeederResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan alertFeederResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	input, err := r.buildInput(plan)
	if err != nil {
		resp.Diagnostics.AddError("Error building alert feeder input", err.Error())
		return
	}
	_, err = r.client.CreateAlertFeeder(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Error creating alert feeder", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.Name.ValueString())
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *alertFeederResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state alertFeederResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.client.GetAlertFeeder(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading alert feeder", err.Error())
		return
	}
	if err := r.mapToState(result, &state); err != nil {
		resp.Diagnostics.AddError("Error mapping alert feeder state", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *alertFeederResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan alertFeederResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state alertFeederResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input, err := r.buildInput(plan)
	if err != nil {
		resp.Diagnostics.AddError("Error building alert feeder input", err.Error())
		return
	}
	result, err := r.client.UpdateAlertFeeder(ctx, state.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Error updating alert feeder", err.Error())
		return
	}
	if err := r.mapToState(result, &plan); err != nil {
		resp.Diagnostics.AddError("Error mapping alert feeder state", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *alertFeederResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state alertFeederResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteAlertFeeder(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting alert feeder", err.Error())
	}
}

func (r *alertFeederResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
