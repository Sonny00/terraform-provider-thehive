package provider

import (
	"context"
	"fmt"

	"github.com/Sonny00/terraform-provider-thehive/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &caseTemplateDataSource{}

func NewCaseTemplateDataSource() datasource.DataSource {
	return &caseTemplateDataSource{}
}

type caseTemplateDataSource struct {
	client *client.Client
}

type caseTemplateDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Description types.String `tfsdk:"description"`
	Severity    types.Int64  `tfsdk:"severity"`
	TLP         types.Int64  `tfsdk:"tlp"`
	PAP         types.Int64  `tfsdk:"pap"`
}

func (d *caseTemplateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_case_template"
}

func (d *caseTemplateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing TheHive case template by name.",
		Attributes: map[string]schema.Attribute{
			"id":           schema.StringAttribute{Description: "The internal ID of the case template.", Computed: true},
			"name":         schema.StringAttribute{Description: "The name of the case template to look up.", Required: true},
			"display_name": schema.StringAttribute{Description: "The display name of the case template.", Computed: true},
			"description":  schema.StringAttribute{Description: "The description of the case template.", Computed: true},
			"severity":     schema.Int64Attribute{Description: "The default severity level (1-4).", Computed: true},
			"tlp":          schema.Int64Attribute{Description: "The default TLP level.", Computed: true},
			"pap":          schema.Int64Attribute{Description: "The default PAP level.", Computed: true},
		},
	}
}

func (d *caseTemplateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}
	d.client = c
}

func (d *caseTemplateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config caseTemplateDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.GetCaseTemplate(ctx, config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading case template", err.Error())
		return
	}

	config.ID = types.StringValue(result.ID)
	config.Name = types.StringValue(result.Name)
	config.DisplayName = types.StringValue(result.DisplayName)
	config.Description = types.StringValue(result.Description)
	config.Severity = types.Int64Value(int64(result.Severity))
	config.TLP = types.Int64Value(int64(result.TLP))
	config.PAP = types.Int64Value(int64(result.PAP))
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
