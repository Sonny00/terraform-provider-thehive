package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Sonny00/terraform-provider-thehive/internal/client"
)

var _ datasource.DataSource = &ProfileDataSource{}

type ProfileDataSource struct{ client *client.Client }

type profileDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Permissions types.Set    `tfsdk:"permissions"`
	IsAdmin     types.Bool   `tfsdk:"is_admin"`
}

func NewProfileDataSource() datasource.DataSource { return &ProfileDataSource{} }

func (d *ProfileDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_profile"
}

func (d *ProfileDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing TheHive profile by name.",
		Attributes: map[string]schema.Attribute{
			"id":   schema.StringAttribute{Computed: true},
			"name": schema.StringAttribute{Description: "Profile name to look up.", Required: true},
			"permissions": schema.SetAttribute{
				Description: "Permissions granted by this profile.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"is_admin": schema.BoolAttribute{
				Description: "Whether this is an admin profile.",
				Computed:    true,
			},
		},
	}
}

func (d *ProfileDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProfileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config profileDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p, err := d.client.GetProfile(ctx, config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Profile not found", fmt.Sprintf("No profile named %q: %s", config.Name.ValueString(), err.Error()))
		return
	}

	config.ID = types.StringValue(p.ID)
	config.Name = types.StringValue(p.Name)
	config.IsAdmin = types.BoolValue(p.IsAdmin)
	permsList, diags := types.SetValueFrom(ctx, types.StringType, p.Permissions)
	resp.Diagnostics.Append(diags...)
	config.Permissions = permsList
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}