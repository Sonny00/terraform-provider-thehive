package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Sonny00/terraform-provider-thehive/internal/client"
)

var _ datasource.DataSource = &OrganisationDataSource{}

type OrganisationDataSource struct{ client *client.Client }

type organisationDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func NewOrganisationDataSource() datasource.DataSource { return &OrganisationDataSource{} }

func (d *OrganisationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organisation"
}

func (d *OrganisationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing TheHive organisation by name.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true},
			"name":        schema.StringAttribute{Description: "Organisation name to look up.", Required: true},
			"description": schema.StringAttribute{Computed: true},
		},
	}
}

func (d *OrganisationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected type", fmt.Sprintf("got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *OrganisationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config organisationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, err := d.client.GetOrganisation(ctx, config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Organisation not found", fmt.Sprintf("No organisation named %q: %s", config.Name.ValueString(), err.Error()))
		return
	}

	config.ID = types.StringValue(org.ID)
	config.Name = types.StringValue(org.Name)
	config.Description = types.StringValue(org.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}