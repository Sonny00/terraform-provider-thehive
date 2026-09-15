package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Sonny00/terraform-provider-thehive/internal/client"
)

var (
	_ resource.Resource                = &PageTemplateResource{}
	_ resource.ResourceWithImportState = &PageTemplateResource{}
)

type PageTemplateResource struct{ client *client.Client }

type pageTemplateResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Title    types.String `tfsdk:"title"`
	Content  types.String `tfsdk:"content"`
	Category types.String `tfsdk:"category"`
	Order    types.Int64  `tfsdk:"order"`
}

func NewPageTemplateResource() resource.Resource { return &PageTemplateResource{} }

func (r *PageTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_page_template"
}

func (r *PageTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TheHive page template for cases.",
		Attributes: map[string]schema.Attribute{
			"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"title":    schema.StringAttribute{Description: "Page title.", Required: true},
			"content":  schema.StringAttribute{Description: "Page body (Markdown).", Optional: true, Computed: true},
			"category": schema.StringAttribute{Description: "Page category.", Optional: true, Computed: true},
			"order":    schema.Int64Attribute{Description: "Display order.", Optional: true, Computed: true, Default: int64default.StaticInt64(0)},
		},
	}
}

func (r *PageTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil { return }
	c, ok := req.ProviderData.(*client.Client)
	if !ok { resp.Diagnostics.AddError("Unexpected type", fmt.Sprintf("got: %T", req.ProviderData)); return }
	r.client = c
}

func (r *PageTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan pageTemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() { return }
	pt, err := r.client.CreatePageTemplate(ctx, client.PageTemplateInput{
		Title: plan.Title.ValueString(), Content: plan.Content.ValueString(),
		Category: plan.Category.ValueString(), Order: int(plan.Order.ValueInt64()),
	})
	if err != nil { resp.Diagnostics.AddError("Error creating page template", err.Error()); return }
	plan.ID = types.StringValue(pt.ID); plan.Title = types.StringValue(pt.Title)
	plan.Content = types.StringValue(pt.Content); plan.Category = types.StringValue(pt.Category)
	plan.Order = types.Int64Value(int64(pt.Order))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *PageTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state pageTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() { return }
	pt, err := r.client.GetPageTemplate(ctx, state.ID.ValueString())
	if err != nil { resp.Diagnostics.AddError("Error reading page template", err.Error()); return }
	state.Title = types.StringValue(pt.Title); state.Content = types.StringValue(pt.Content)
	state.Category = types.StringValue(pt.Category); state.Order = types.Int64Value(int64(pt.Order))
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *PageTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan pageTemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() { return }
	pt, err := r.client.UpdatePageTemplate(ctx, plan.ID.ValueString(), client.PageTemplateInput{
		Title: plan.Title.ValueString(), Content: plan.Content.ValueString(),
		Category: plan.Category.ValueString(), Order: int(plan.Order.ValueInt64()),
	})
	if err != nil { resp.Diagnostics.AddError("Error updating page template", err.Error()); return }
	plan.Title = types.StringValue(pt.Title); plan.Content = types.StringValue(pt.Content)
	plan.Category = types.StringValue(pt.Category); plan.Order = types.Int64Value(int64(pt.Order))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *PageTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state pageTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() { return }
	if err := r.client.DeletePageTemplate(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting page template", err.Error())
	}
}

func (r *PageTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
