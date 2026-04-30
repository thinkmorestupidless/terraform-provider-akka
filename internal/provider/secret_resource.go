package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thinkmorestupidless/terraform-provider-akka/internal/akka"
)

var _ resource.Resource = &SecretResource{}
var _ resource.ResourceWithImportState = &SecretResource{}

type SecretResource struct {
	client *akka.AkkaClient
}

type secretResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Project           types.String `tfsdk:"project"`
	Type              types.String `tfsdk:"type"`
	Value             types.String `tfsdk:"value"`
	PublicKey         types.String `tfsdk:"public_key"`
	PrivateKey        types.String `tfsdk:"private_key"`
	Certificate       types.String `tfsdk:"certificate"`
	CACertificate     types.String `tfsdk:"ca_certificate"`
	ExternalProvider  types.String `tfsdk:"external_provider"`
	ExternalReference types.String `tfsdk:"external_reference"`
	Organization      types.String `tfsdk:"organization"`
}

func NewSecretResource() resource.Resource {
	return &SecretResource{}
}

func (r *SecretResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (r *SecretResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replaceOnChange := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		Description: "Manages an Akka secret. Secret values are sensitive and never returned by the platform after creation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:      true,
				PlanModifiers: replaceOnChange,
			},
			"project": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Required:      true,
				Description:   "Secret type: symmetric, asymmetric, generic, tls, tls-ca.",
				PlanModifiers: replaceOnChange,
			},
			"value": schema.StringAttribute{
				Optional:      true,
				Sensitive:     true,
				Description:   "Secret value for symmetric and generic types.",
				PlanModifiers: replaceOnChange,
			},
			"public_key": schema.StringAttribute{
				Optional:      true,
				Description:   "Public key for asymmetric type (PEM).",
				PlanModifiers: replaceOnChange,
			},
			"private_key": schema.StringAttribute{
				Optional:      true,
				Sensitive:     true,
				Description:   "Private key for asymmetric and tls types (PEM).",
				PlanModifiers: replaceOnChange,
			},
			"certificate": schema.StringAttribute{
				Optional:      true,
				Description:   "Certificate for tls type (PEM).",
				PlanModifiers: replaceOnChange,
			},
			"ca_certificate": schema.StringAttribute{
				Optional:      true,
				Description:   "CA certificate for tls-ca type (PEM).",
				PlanModifiers: replaceOnChange,
			},
			"external_provider": schema.StringAttribute{
				Optional:    true,
				Description: "External provider: aws, azure, gcp.",
			},
			"external_reference": schema.StringAttribute{
				Optional:    true,
				Description: "External secret reference (ARN, URI, or name).",
			},
			"organization": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *SecretResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		&secretTypeValidator{},
	}
}

func (r *SecretResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*akka.AkkaClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("expected *akka.AkkaClient, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *SecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan secretResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := akka.SecretCreateRequest{
		Name:              plan.Name.ValueString(),
		Project:           plan.Project.ValueString(),
		Type:              plan.Type.ValueString(),
		Value:             plan.Value.ValueString(),
		PublicKey:         plan.PublicKey.ValueString(),
		PrivateKey:        plan.PrivateKey.ValueString(),
		Certificate:       plan.Certificate.ValueString(),
		CACertificate:     plan.CACertificate.ValueString(),
		ExternalProvider:  plan.ExternalProvider.ValueString(),
		ExternalReference: plan.ExternalReference.ValueString(),
	}
	if err := r.client.CreateSecret(ctx, createReq); err != nil {
		resp.Diagnostics.AddError("Create Secret Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.Project.ValueString() + "/" + plan.Name.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state secretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.GetSecret(ctx, state.Name.ValueString(), state.Project.ValueString())
	if err != nil {
		if akka.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Secret Failed", err.Error())
		return
	}
	// Values are not returned by the platform — state is authoritative for value fields.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SecretResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All fields have RequiresReplace — Update should never be called.
	resp.Diagnostics.AddError("Update Not Supported", "All secret fields force resource recreation.")
}

func (r *SecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state secretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteSecret(ctx, state.Name.ValueString(), state.Project.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete Secret Failed", err.Error())
	}
}

func (r *SecretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid Import ID", "Expected format: <project>/<name>")
		return
	}
	secret, err := r.client.GetSecret(ctx, parts[1], parts[0])
	if err != nil {
		resp.Diagnostics.AddError("Import Secret Failed", err.Error())
		return
	}
	state := secretResourceModel{
		ID:      types.StringValue(parts[0] + "/" + parts[1]),
		Name:    types.StringValue(secret.Name),
		Project: types.StringValue(secret.Project),
		Type:    types.StringValue(secret.Type),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// secretTypeValidator validates that the correct fields are set for each secret type.
type secretTypeValidator struct{}

func (v *secretTypeValidator) Description(_ context.Context) string {
	return "Validates that required fields are set for the specified secret type."
}

func (v *secretTypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *secretTypeValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config secretResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	secretType := config.Type.ValueString()
	switch secretType {
	case "symmetric", "generic":
		if config.Value.IsNull() || config.Value.IsUnknown() || config.Value.ValueString() == "" {
			resp.Diagnostics.AddError(
				"Missing value",
				fmt.Sprintf("The 'value' field is required for secret type %q.", secretType),
			)
		}
	case "asymmetric":
		if config.PublicKey.ValueString() == "" || config.PrivateKey.ValueString() == "" {
			resp.Diagnostics.AddError("Missing Keys", "Both 'public_key' and 'private_key' are required for asymmetric secrets.")
		}
	case "tls":
		if config.Certificate.ValueString() == "" || config.PrivateKey.ValueString() == "" {
			resp.Diagnostics.AddError("Missing TLS Fields", "Both 'certificate' and 'private_key' are required for tls secrets.")
		}
	case "tls-ca":
		if config.CACertificate.ValueString() == "" {
			resp.Diagnostics.AddError("Missing CA Certificate", "'ca_certificate' is required for tls-ca secrets.")
		}
	}
}
