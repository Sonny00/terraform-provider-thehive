package provider

import (
	"context"
	"fmt"

	"github.com/Sonny00/terraform-provider-thehive/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &emailIntakeResource{}
	_ resource.ResourceWithImportState = &emailIntakeResource{}
)

type emailIntakeResource struct {
	client *client.Client
}

type emailIntakeBasicAuthModel struct {
	Password types.String `tfsdk:"password"`
}

type emailIntakeOAuth2Model struct {
	ClientID          types.String `tfsdk:"client_id"`
	TenantID          types.String `tfsdk:"tenant_id"`
	Secret            types.String `tfsdk:"secret"`
	Authority         types.String `tfsdk:"authority"`
	Scopes            types.List   `tfsdk:"scopes"`
	RedirectURI       types.String `tfsdk:"redirect_uri"`
	AuthorizationCode types.String `tfsdk:"authorization_code"`
}

type emailIntakeCredentialModel struct {
	Email     types.String               `tfsdk:"email"`
	BasicAuth *emailIntakeBasicAuthModel `tfsdk:"basic_auth"`
	OAuth2    *emailIntakeOAuth2Model    `tfsdk:"oauth2"`
}

type emailIntakeMailboxModel struct {
	ProviderName types.String                `tfsdk:"provider_name"`
	Credential   *emailIntakeCredentialModel `tfsdk:"credential"`
	Inbox        types.String                `tfsdk:"inbox"`
	Archive      types.String                `tfsdk:"archive"`
	MarkAsRead   types.Bool                  `tfsdk:"mark_as_read"`
}

type emailIntakeAlertPropertiesModel struct {
	Type   types.String `tfsdk:"type"`
	Source types.String `tfsdk:"source"`
	Tags   types.List   `tfsdk:"tags"`
}

type emailIntakeResourceModel struct {
	ID              types.String                     `tfsdk:"id"`
	Name            types.String                     `tfsdk:"name"`
	Enabled         types.Bool                       `tfsdk:"enabled"`
	Organisations   types.List                       `tfsdk:"organisations"`
	Mailbox         *emailIntakeMailboxModel         `tfsdk:"mailbox"`
	AlertProperties *emailIntakeAlertPropertiesModel `tfsdk:"alert_properties"`
}

func NewEmailIntakeResource() resource.Resource {
	return &emailIntakeResource{}
}

func (r *emailIntakeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_email_intake"
}

func (r *emailIntakeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an email intake connector in TheHive 5.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "The internal ID of the email intake connector.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description: "A unique name for this email intake connector.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether this connector is active.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"organisations": schema.ListAttribute{
				Description: "TheHive organisation names that receive alerts from this connector.",
				Required:    true,
				ElementType: types.StringType,
			},
			"mailbox": schema.SingleNestedAttribute{
				Description: "Mailbox connection configuration.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"provider_name": schema.StringAttribute{
						Description: "Mail provider: imap, google-workspace, office365, or MSGraph365.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("imap", "google-workspace", "office365", "MSGraph365"),
						},
					},
					"inbox": schema.StringAttribute{
						Description: "Name of the inbox folder to monitor.",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString("Inbox"),
					},
					"archive": schema.StringAttribute{
						Description: "Name of the archive folder for processed emails.",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString("Archive"),
					},
					"mark_as_read": schema.BoolAttribute{
						Description: "Whether to mark processed emails as read.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"credential": schema.SingleNestedAttribute{
						Description: "Authentication credentials for the mailbox.",
						Required:    true,
						Attributes: map[string]schema.Attribute{
							"email": schema.StringAttribute{
								Description: "Email address of the mailbox.",
								Required:    true,
							},
							"basic_auth": schema.SingleNestedAttribute{
								Description: "Basic authentication (IMAP password).",
								Optional:    true,
								Attributes: map[string]schema.Attribute{
									"password": schema.StringAttribute{
										Description: "Mailbox password.",
										Required:    true,
										Sensitive:   true,
									},
								},
							},
							"oauth2": schema.SingleNestedAttribute{
								Description: "OAuth2 authentication credentials.",
								Optional:    true,
								Attributes: map[string]schema.Attribute{
									"client_id": schema.StringAttribute{
										Description: "OAuth2 client ID.",
										Required:    true,
										Sensitive:   true,
									},
									"tenant_id": schema.StringAttribute{
										Description: "Azure AD tenant ID, when required by the provider.",
										Optional:    true,
									},
									"secret": schema.StringAttribute{
										Description: "OAuth2 client secret.",
										Required:    true,
										Sensitive:   true,
									},
									"authority": schema.StringAttribute{
										Description: "OAuth2 authority URL.",
										Optional:    true,
									},
									"scopes": schema.ListAttribute{
										Description: "OAuth2 scopes.",
										Optional:    true,
										ElementType: types.StringType,
									},
									"redirect_uri": schema.StringAttribute{
										Description: "OAuth2 redirect URI.",
										Optional:    true,
									},
									"authorization_code": schema.StringAttribute{
										Description: "OAuth2 authorization code.",
										Optional:    true,
										Sensitive:   true,
									},
								},
							},
						},
					},
				},
			},
			"alert_properties": schema.SingleNestedAttribute{
				Description: "Properties applied to alerts created by this connector.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "Alert type (defaults to email-intake).",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString("email-intake"),
					},
					"source": schema.StringAttribute{
						Description: "Alert source. Supports variables such as {name}.",
						Optional:    true,
					},
					"tags": schema.ListAttribute{
						Description: "Tags for created alerts. Supports variables such as {name} and {mailbox.provider.name}.",
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
		},
	}
}

func (r *emailIntakeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *emailIntakeResource) buildInput(ctx context.Context, plan emailIntakeResourceModel) (*client.EmailIntakeInput, error) {
	var organisations []string
	if diags := plan.Organisations.ElementsAs(ctx, &organisations, false); diags.HasError() {
		return nil, fmt.Errorf("reading organisations from plan: %s", diags.Errors()[0].Detail())
	}

	input := &client.EmailIntakeInput{
		Name:          plan.Name.ValueString(),
		Enabled:       plan.Enabled.ValueBool(),
		Organisations: organisations,
	}

	if plan.Mailbox == nil || plan.Mailbox.Credential == nil {
		return nil, fmt.Errorf("mailbox and mailbox credential must be configured")
	}
	input.Mailbox = client.EmailIntakeMailbox{
		Provider:   client.EmailIntakeProvider{Name: plan.Mailbox.ProviderName.ValueString()},
		Inbox:      plan.Mailbox.Inbox.ValueString(),
		Archive:    plan.Mailbox.Archive.ValueString(),
		MarkAsRead: plan.Mailbox.MarkAsRead.ValueBool(),
		Kind:       "api",
		Credential: client.EmailIntakeCredential{
			Email: plan.Mailbox.Credential.Email.ValueString(),
		},
	}

	if auth := plan.Mailbox.Credential.BasicAuth; auth != nil {
		input.Mailbox.Credential.BasicAuth = &client.EmailIntakeBasicAuth{
			Password: auth.Password.ValueString(),
		}
	}
	if auth := plan.Mailbox.Credential.OAuth2; auth != nil {
		oauth2 := &client.EmailIntakeOAuth2{
			ClientID:          auth.ClientID.ValueString(),
			Secret:            auth.Secret.ValueString(),
			TenantID:          auth.TenantID.ValueString(),
			Authority:         auth.Authority.ValueString(),
			RedirectURI:       auth.RedirectURI.ValueString(),
			AuthorizationCode: auth.AuthorizationCode.ValueString(),
		}
		if !auth.Scopes.IsNull() && !auth.Scopes.IsUnknown() {
			var scopes []string
			if diags := auth.Scopes.ElementsAs(ctx, &scopes, false); diags.HasError() {
				return nil, fmt.Errorf("reading OAuth2 scopes from plan: %s", diags.Errors()[0].Detail())
			}
			oauth2.Scopes = scopes
		}
		input.Mailbox.Credential.OAuth2 = oauth2
	}

	if properties := plan.AlertProperties; properties != nil {
		alertProperties := &client.EmailIntakeAlertProperties{
			Type:   properties.Type.ValueString(),
			Source: properties.Source.ValueString(),
		}
		if !properties.Tags.IsNull() && !properties.Tags.IsUnknown() {
			var tags []string
			if diags := properties.Tags.ElementsAs(ctx, &tags, false); diags.HasError() {
				return nil, fmt.Errorf("reading alert tags from plan: %s", diags.Errors()[0].Detail())
			}
			alertProperties.Tags = tags
		}
		input.AlertProperties = alertProperties
	}

	return input, nil
}

func stringsToList(values []string) types.List {
	elements := make([]attr.Value, len(values))
	for i, value := range values {
		elements[i] = types.StringValue(value)
	}
	list, _ := types.ListValue(types.StringType, elements)
	return list
}

func (r *emailIntakeResource) mapToState(result *client.EmailIntake, state *emailIntakeResourceModel) {
	state.ID = types.StringValue(result.ID)
	state.Name = types.StringValue(result.Name)
	state.Enabled = types.BoolValue(result.Enabled)
	state.Organisations = stringsToList(result.Organisations)

	mailbox := &emailIntakeMailboxModel{
		ProviderName: types.StringValue(result.Mailbox.Provider.Name),
		Inbox:        types.StringValue(result.Mailbox.Inbox),
		Archive:      types.StringValue(result.Mailbox.Archive),
		MarkAsRead:   types.BoolValue(result.Mailbox.MarkAsRead),
		Credential: &emailIntakeCredentialModel{
			Email: types.StringValue(result.Mailbox.Credential.Email),
		},
	}

	if auth := result.Mailbox.Credential.BasicAuth; auth != nil {
		mailbox.Credential.BasicAuth = &emailIntakeBasicAuthModel{
			Password: types.StringValue(auth.Password),
		}
	}
	if auth := result.Mailbox.Credential.OAuth2; auth != nil {
		scopes := types.ListNull(types.StringType)
		if len(auth.Scopes) > 0 {
			scopes = stringsToList(auth.Scopes)
		}
		mailbox.Credential.OAuth2 = &emailIntakeOAuth2Model{
			ClientID:          types.StringValue(auth.ClientID),
			TenantID:          types.StringValue(auth.TenantID),
			Secret:            types.StringValue(auth.Secret),
			Authority:         types.StringValue(auth.Authority),
			Scopes:            scopes,
			RedirectURI:       types.StringValue(auth.RedirectURI),
			AuthorizationCode: types.StringValue(auth.AuthorizationCode),
		}
	}
	state.Mailbox = mailbox

	if properties := result.AlertProperties; properties != nil {
		tags := types.ListNull(types.StringType)
		if len(properties.Tags) > 0 {
			tags = stringsToList(properties.Tags)
		}
		state.AlertProperties = &emailIntakeAlertPropertiesModel{
			Type:   types.StringValue(properties.Type),
			Source: types.StringValue(properties.Source),
			Tags:   tags,
		}
	} else {
		state.AlertProperties = nil
	}
}

func (r *emailIntakeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan emailIntakeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, err := r.buildInput(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error building email intake input", err.Error())
		return
	}
	result, err := r.client.CreateEmailIntake(ctx, *input)
	if err != nil {
		resp.Diagnostics.AddError("Error creating email intake", err.Error())
		return
	}

	r.mapToState(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *emailIntakeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state emailIntakeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetEmailIntake(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading email intake", err.Error())
		return
	}
	r.mapToState(result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *emailIntakeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan emailIntakeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state emailIntakeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, err := r.buildInput(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error building email intake input", err.Error())
		return
	}
	result, err := r.client.UpdateEmailIntake(ctx, state.ID.ValueString(), *input)
	if err != nil {
		resp.Diagnostics.AddError("Error updating email intake", err.Error())
		return
	}
	r.mapToState(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *emailIntakeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state emailIntakeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteEmailIntake(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting email intake", err.Error())
	}
}

func (r *emailIntakeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	result, err := r.client.GetEmailIntake(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing email intake", err.Error())
		return
	}
	var state emailIntakeResourceModel
	r.mapToState(result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
