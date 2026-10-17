package provider

import (
	"context"
	"fmt"

	"github.com/Sonny00/terraform-provider-thehive/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &customFieldDataSource{}

func NewCustomFieldDataSource() datasource.DataSource {
	return &customFieldDataSource{}
}

type customFieldDataSource struct {
	client *client.Client
}

type customFieldDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Mandatory   types.Bool   `tfsdk:"mandatory"`
	Group       types.String `tfsdk:"group"`
}

func (d *customFieldDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_field"
}

func (d *customFieldDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing TheHive custom field by name.",
		Attributes: map[string]schema.Attribute{
			"id":           schema.StringAttribute{Description: "The internal ID of the custom field.", Computed: true},
			"name":         schema.StringAttribute{Description: "The technical name of the custom field to look up.", Required: true},
			"display_name": schema.StringAttribute{Description: "The display name of the custom field.", Computed: true},
			"description":  schema.StringAttribute{Description: "The description of the custom field.", Computed: true},
			"type":         schema.StringAttribute{Description: "The type of the custom field (string, integer, boolean, date, float).", Computed: true},
			"mandatory":    schema.BoolAttribute{Description: "Whether the custom field is mandatory.", Computed: true},
			"group":        schema.StringAttribute{Description: "The group this custom field belongs to.", Computed: true},
		},
	}
}

func (d *customFieldDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *customFieldDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config customFieldDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.GetCustomField(ctx, config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading custom field", err.Error())
		return
	}

	config.ID = types.StringValue(result.ID)
	config.Name = types.StringValue(result.Name)
	config.DisplayName = types.StringValue(result.DisplayName)
	config.Description = types.StringValue(result.Description)
	config.Type = types.StringValue(result.Type)
	config.Mandatory = types.BoolValue(result.Mandatory)
	config.Group = types.StringValue(result.Group)
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
