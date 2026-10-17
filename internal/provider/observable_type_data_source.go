package provider

import (
	"context"
	"fmt"

	"github.com/Sonny00/terraform-provider-thehive/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &observableTypeDataSource{}

func NewObservableTypeDataSource() datasource.DataSource {
	return &observableTypeDataSource{}
}

type observableTypeDataSource struct {
	client *client.Client
}

type observableTypeDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	IsAttachment types.Bool   `tfsdk:"is_attachment"`
}

func (d *observableTypeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_observable_type"
}

func (d *observableTypeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing TheHive observable type by name.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Description: "The internal ID of the observable type.", Computed: true},
			"name":          schema.StringAttribute{Description: "The name of the observable type to look up.", Required: true},
			"is_attachment": schema.BoolAttribute{Description: "Whether this observable type is a file attachment.", Computed: true},
		},
	}
}

func (d *observableTypeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *observableTypeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config observableTypeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.GetObservableType(ctx, config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading observable type", err.Error())
		return
	}

	config.ID = types.StringValue(result.ID)
	config.Name = types.StringValue(result.Name)
	config.IsAttachment = types.BoolValue(result.IsAttachment)
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
