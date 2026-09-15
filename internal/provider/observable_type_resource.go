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
	_ resource.Resource                = &ObservableTypeResource{}
	_ resource.ResourceWithImportState = &ObservableTypeResource{}
)

type ObservableTypeResource struct{ client *client.Client }

type observableTypeResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	IsAttachment types.Bool   `tfsdk:"is_attachment"`
}

func NewObservableTypeResource() resource.Resource { return &ObservableTypeResource{} }

func (r *ObservableTypeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_observable_type"
}

func (r *ObservableTypeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TheHive observable type.",
		Attributes: map[string]schema.Attribute{
			"id":   schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name": schema.StringAttribute{Description: "Observable type name (e.g. ip, domain, hash).", Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"is_attachment": schema.BoolAttribute{Description: "Whether this type represents a file attachment.", Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
		},
	}
}

func (r *ObservableTypeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil { return }
	c, ok := req.ProviderData.(*client.Client)
	if !ok { resp.Diagnostics.AddError("Unexpected type", fmt.Sprintf("got: %T", req.ProviderData)); return }
	r.client = c
}

func (r *ObservableTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan observableTypeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() { return }
	ot, err := r.client.CreateObservableType(ctx, client.ObservableTypeInput{Name: plan.Name.ValueString(), IsAttachment: plan.IsAttachment.ValueBool()})
	if err != nil { resp.Diagnostics.AddError("Error creating observable type", err.Error()); return }
	plan.ID = types.StringValue(ot.ID); plan.Name = types.StringValue(ot.Name); plan.IsAttachment = types.BoolValue(ot.IsAttachment)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ObservableTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state observableTypeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() { return }
	ot, err := r.client.GetObservableType(ctx, state.ID.ValueString())
	if err != nil { resp.Diagnostics.AddError("Error reading observable type", err.Error()); return }
	state.Name = types.StringValue(ot.Name); state.IsAttachment = types.BoolValue(ot.IsAttachment)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ObservableTypeResource) Update(ctx context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Observable types are immutable. Delete and recreate to change.")
}

func (r *ObservableTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state observableTypeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() { return }
	if err := r.client.DeleteObservableType(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting observable type", err.Error())
	}
}

func (r *ObservableTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
