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
	_ resource.Resource                = &CaseTemplateResource{}
	_ resource.ResourceWithImportState = &CaseTemplateResource{}
)

type CaseTemplateResource struct{ client *client.Client }

type caseTemplateResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	TitlePrefix types.String `tfsdk:"title_prefix"`
	Description types.String `tfsdk:"description"`
	Severity    types.Int64  `tfsdk:"severity"`
	Tags        types.List   `tfsdk:"tags"`
	TLP         types.Int64  `tfsdk:"tlp"`
	PAP         types.Int64  `tfsdk:"pap"`
	Summary     types.String `tfsdk:"summary"`
}

func NewCaseTemplateResource() resource.Resource { return &CaseTemplateResource{} }

func (r *CaseTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_case_template"
}

func (r *CaseTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TheHive case template.",
		Attributes: map[string]schema.Attribute{
			"id":           schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":         schema.StringAttribute{Description: "Template internal name.", Required: true},
			"display_name": schema.StringAttribute{Description: "Display name.", Optional: true, Computed: true},
			"title_prefix": schema.StringAttribute{Description: "Prefix prepended to case titles.", Optional: true, Computed: true},
			"description":  schema.StringAttribute{Description: "Template description.", Optional: true, Computed: true},
			"severity":     schema.Int64Attribute{Description: "Default severity (1=low, 2=medium, 3=high, 4=critical).", Optional: true, Computed: true, Default: int64default.StaticInt64(2)},
			"tags":         schema.ListAttribute{Description: "Default tags.", Optional: true, ElementType: types.StringType},
			"tlp":          schema.Int64Attribute{Description: "Default TLP (0=clear, 1=green, 2=amber, 3=amber+strict, 4=red).", Optional: true, Computed: true, Default: int64default.StaticInt64(2)},
			"pap":          schema.Int64Attribute{Description: "Default PAP (0=clear, 1=green, 2=amber, 3=red).", Optional: true, Computed: true, Default: int64default.StaticInt64(2)},
			"summary":      schema.StringAttribute{Description: "Default case summary.", Optional: true, Computed: true},
		},
	}
}

func (r *CaseTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil { return }
	c, ok := req.ProviderData.(*client.Client)
	if !ok { resp.Diagnostics.AddError("Unexpected type", fmt.Sprintf("got: %T", req.ProviderData)); return }
	r.client = c
}

func (r *CaseTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan caseTemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() { return }
	var tags []string
	if !plan.Tags.IsNull() { resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...) }
	if resp.Diagnostics.HasError() { return }

	ct, err := r.client.CreateCaseTemplate(ctx, client.CaseTemplateInput{
		Name: plan.Name.ValueString(), DisplayName: plan.DisplayName.ValueString(),
		TitlePrefix: plan.TitlePrefix.ValueString(), Description: plan.Description.ValueString(),
		Severity: int(plan.Severity.ValueInt64()), Tags: tags,
		TLP: int(plan.TLP.ValueInt64()), PAP: int(plan.PAP.ValueInt64()),
		Summary: plan.Summary.ValueString(),
	})
	if err != nil { resp.Diagnostics.AddError("Error creating case template", err.Error()); return }

	plan.ID = types.StringValue(ct.ID); plan.Name = types.StringValue(ct.Name)
	plan.DisplayName = types.StringValue(ct.DisplayName); plan.TitlePrefix = types.StringValue(ct.TitlePrefix)
	plan.Description = types.StringValue(ct.Description); plan.Severity = types.Int64Value(int64(ct.Severity))
	plan.TLP = types.Int64Value(int64(ct.TLP)); plan.PAP = types.Int64Value(int64(ct.PAP))
	plan.Summary = types.StringValue(ct.Summary)
	if len(ct.Tags) > 0 { tl, d := types.ListValueFrom(ctx, types.StringType, ct.Tags); resp.Diagnostics.Append(d...); plan.Tags = tl }
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CaseTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state caseTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() { return }
	ct, err := r.client.GetCaseTemplate(ctx, state.ID.ValueString())
	if err != nil { resp.Diagnostics.AddError("Error reading case template", err.Error()); return }
	state.Name = types.StringValue(ct.Name); state.DisplayName = types.StringValue(ct.DisplayName)
	state.TitlePrefix = types.StringValue(ct.TitlePrefix); state.Description = types.StringValue(ct.Description)
	state.Severity = types.Int64Value(int64(ct.Severity)); state.TLP = types.Int64Value(int64(ct.TLP))
	state.PAP = types.Int64Value(int64(ct.PAP)); state.Summary = types.StringValue(ct.Summary)
	if len(ct.Tags) > 0 { tl, d := types.ListValueFrom(ctx, types.StringType, ct.Tags); resp.Diagnostics.Append(d...); state.Tags = tl }
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *CaseTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan caseTemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() { return }
	var tags []string
	if !plan.Tags.IsNull() { resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...) }
	if resp.Diagnostics.HasError() { return }
	ct, err := r.client.UpdateCaseTemplate(ctx, plan.ID.ValueString(), client.CaseTemplateInput{
		Name: plan.Name.ValueString(), DisplayName: plan.DisplayName.ValueString(),
		TitlePrefix: plan.TitlePrefix.ValueString(), Description: plan.Description.ValueString(),
		Severity: int(plan.Severity.ValueInt64()), Tags: tags,
		TLP: int(plan.TLP.ValueInt64()), PAP: int(plan.PAP.ValueInt64()), Summary: plan.Summary.ValueString(),
	})
	if err != nil { resp.Diagnostics.AddError("Error updating case template", err.Error()); return }
	plan.Name = types.StringValue(ct.Name); plan.DisplayName = types.StringValue(ct.DisplayName)
	plan.Description = types.StringValue(ct.Description); plan.Summary = types.StringValue(ct.Summary)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CaseTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state caseTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() { return }
	if err := r.client.DeleteCaseTemplate(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting case template", err.Error())
	}
}

func (r *CaseTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
