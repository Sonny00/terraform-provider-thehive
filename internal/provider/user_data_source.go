package provider

import (
	"context"
	"fmt"

	"github.com/Sonny00/terraform-provider-thehive/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &userDataSource{}

func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

type userDataSource struct {
	client *client.Client
}

type userDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	Login   types.String `tfsdk:"login"`
	Name    types.String `tfsdk:"name"`
	Profile types.String `tfsdk:"profile"`
	Type    types.String `tfsdk:"type"`
	Locked  types.Bool   `tfsdk:"locked"`
}

func (d *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *userDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing TheHive user by login.",
		Attributes: map[string]schema.Attribute{
			"id":      schema.StringAttribute{Description: "The internal ID of the user.", Computed: true},
			"login":   schema.StringAttribute{Description: "The login (email) of the user to look up.", Required: true},
			"name":    schema.StringAttribute{Description: "The display name of the user.", Computed: true},
			"profile": schema.StringAttribute{Description: "The profile assigned to the user.", Computed: true},
			"type":    schema.StringAttribute{Description: "The type of the user (Normal or Service).", Computed: true},
			"locked":  schema.BoolAttribute{Description: "Whether the user account is locked.", Computed: true},
		},
	}
}

func (d *userDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config userDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.GetUser(ctx, config.Login.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading user", err.Error())
		return
	}

	config.ID = types.StringValue(result.ID)
	config.Login = types.StringValue(result.Login)
	config.Name = types.StringValue(result.Name)
	config.Profile = types.StringValue(result.Profile)
	config.Type = types.StringValue(result.Type)
	config.Locked = types.BoolValue(result.Locked)
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
