package provider

import (
	"context"
	"fmt"

	"github.com/Sonny00/terraform-provider-thehive/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &taxonomyResource{}
	_ resource.ResourceWithImportState = &taxonomyResource{}
)

func NewTaxonomyResource() resource.Resource {
	return &taxonomyResource{}
}

type taxonomyResource struct {
	client *client.Client
}

type taxonomyEntryModel struct {
	Value          types.String `tfsdk:"value"`
	Expanded       types.String `tfsdk:"expanded"`
	Colour         types.String `tfsdk:"colour"`
	Description    types.String `tfsdk:"description"`
	NumericalValue types.Int64  `tfsdk:"numerical_value"`
}

type taxonomyValueModel struct {
	Predicate types.String         `tfsdk:"predicate"`
	Entry     []taxonomyEntryModel `tfsdk:"entry"`
}

type taxonomyPredicateModel struct {
	Value       types.String `tfsdk:"value"`
	Expanded    types.String `tfsdk:"expanded"`
	Exclusive   types.Bool   `tfsdk:"exclusive"`
	Description types.String `tfsdk:"description"`
	Colour      types.String `tfsdk:"colour"`
}

type taxonomyResourceModel struct {
	ID          types.String             `tfsdk:"id"`
	Namespace   types.String             `tfsdk:"namespace"`
	Description types.String             `tfsdk:"description"`
	Version     types.Int64              `tfsdk:"version"`
	Exclusive   types.Bool               `tfsdk:"exclusive"`
	Predicates  []taxonomyPredicateModel `tfsdk:"predicates"`
	Values      []taxonomyValueModel     `tfsdk:"values"`
}

func (r *taxonomyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_taxonomy"
}

func (r *taxonomyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Imports a taxonomy into TheHive 5 using the MISP taxonomy format. Changes force replacement.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "The internal ID of the taxonomy.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"namespace": schema.StringAttribute{
				Description:   "The taxonomy namespace (e.g. \"tlp\", \"phishing\").",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"description": schema.StringAttribute{
				Description:   "A description of the taxonomy.",
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"version": schema.Int64Attribute{
				Description:   "The version number of the taxonomy.",
				Optional:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"exclusive": schema.BoolAttribute{
				Description:   "Whether tags from this taxonomy are mutually exclusive.",
				Optional:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
			"predicates": schema.ListNestedAttribute{
				Description:   "List of predicates (tag categories) in this taxonomy.",
				Required:      true,
				PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"value":       schema.StringAttribute{Description: "The predicate value.", Required: true},
						"expanded":    schema.StringAttribute{Description: "The expanded label.", Optional: true},
						"exclusive":   schema.BoolAttribute{Description: "Whether values under this predicate are mutually exclusive.", Optional: true},
						"description": schema.StringAttribute{Description: "Description of the predicate.", Optional: true},
						"colour":      schema.StringAttribute{Description: "Hex colour for this predicate.", Optional: true},
					},
				},
			},
			"values": schema.ListNestedAttribute{
				Description:   "List of values grouped by predicate.",
				Optional:      true,
				PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"predicate": schema.StringAttribute{Description: "The predicate this value group belongs to.", Required: true},
						"entry": schema.ListNestedAttribute{
							Description: "List of entries under this predicate.",
							Required:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"value":           schema.StringAttribute{Description: "The entry value.", Required: true},
									"expanded":        schema.StringAttribute{Description: "The expanded label.", Optional: true},
									"colour":          schema.StringAttribute{Description: "Hex colour for this entry.", Optional: true},
									"description":     schema.StringAttribute{Description: "Description of the entry.", Optional: true},
									"numerical_value": schema.Int64Attribute{Description: "Numerical value for ordering or scoring.", Optional: true},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *taxonomyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *taxonomyResource) buildInput(plan taxonomyResourceModel) client.TaxonomyInput {
	input := client.TaxonomyInput{
		Namespace:   plan.Namespace.ValueString(),
		Description: plan.Description.ValueString(),
		Version:     plan.Version.ValueInt64(),
		Exclusive:   plan.Exclusive.ValueBool(),
	}

	for _, predicate := range plan.Predicates {
		item := client.TaxonomyPredicate{Value: predicate.Value.ValueString()}
		if !predicate.Expanded.IsNull() {
			item.Expanded = predicate.Expanded.ValueString()
		}
		if !predicate.Exclusive.IsNull() {
			item.Exclusive = predicate.Exclusive.ValueBool()
		}
		if !predicate.Description.IsNull() {
			item.Description = predicate.Description.ValueString()
		}
		if !predicate.Colour.IsNull() {
			item.Colour = predicate.Colour.ValueString()
		}
		input.Predicates = append(input.Predicates, item)
	}

	for _, value := range plan.Values {
		item := client.TaxonomyValue{Predicate: value.Predicate.ValueString()}
		for _, entry := range value.Entry {
			valueEntry := client.TaxonomyEntry{Value: entry.Value.ValueString()}
			if !entry.Expanded.IsNull() {
				valueEntry.Expanded = entry.Expanded.ValueString()
			}
			if !entry.Colour.IsNull() {
				valueEntry.Colour = entry.Colour.ValueString()
			}
			if !entry.Description.IsNull() {
				valueEntry.Description = entry.Description.ValueString()
			}
			if !entry.NumericalValue.IsNull() {
				numericalValue := entry.NumericalValue.ValueInt64()
				valueEntry.NumericalValue = &numericalValue
			}
			item.Entry = append(item.Entry, valueEntry)
		}
		input.Values = append(input.Values, item)
	}

	return input
}

func (r *taxonomyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan taxonomyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.CreateTaxonomy(ctx, r.buildInput(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error creating taxonomy", err.Error())
		return
	}

	plan.ID = types.StringValue(result.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *taxonomyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state taxonomyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetTaxonomy(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading taxonomy", err.Error())
		return
	}

	state.ID = types.StringValue(result.ID)
	state.Namespace = types.StringValue(result.Namespace)
	state.Description = types.StringValue(result.Description)
	state.Version = types.Int64Value(result.Version)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *taxonomyResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"Taxonomies cannot be updated through this provider; changes require replacement.",
	)
}

func (r *taxonomyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state taxonomyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteTaxonomy(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting taxonomy", err.Error())
	}
}

func (r *taxonomyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	result, err := r.client.GetTaxonomy(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing taxonomy", err.Error())
		return
	}

	state := taxonomyResourceModel{
		ID:          types.StringValue(result.ID),
		Namespace:   types.StringValue(result.Namespace),
		Description: types.StringValue(result.Description),
		Version:     types.Int64Value(result.Version),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
