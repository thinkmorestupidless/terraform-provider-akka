package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thinkmorestupidless/terraform-provider-akka/internal/akka"
)

var _ resource.Resource = &RoleBindingResource{}
var _ resource.ResourceWithImportState = &RoleBindingResource{}

type RoleBindingResource struct {
	client *akka.AkkaClient
}

type roleBindingResourceModel struct {
	ID           types.String `tfsdk:"id"`
	User         types.String `tfsdk:"user"`
	Role         types.String `tfsdk:"role"`
	Project      types.String `tfsdk:"project"`
	Scope        types.String `tfsdk:"scope"`
	Organization types.String `tfsdk:"organization"`
}

func NewRoleBindingResource() resource.Resource {
	return &RoleBindingResource{}
}

func (r *RoleBindingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_binding"
}

func (r *RoleBindingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replaceOnChange := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		Description: "Manages a role binding assigning a user to a role within a project or organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user": schema.StringAttribute{
				Required:      true,
				Description:   "User email or ID.",
				PlanModifiers: replaceOnChange,
			},
			"role": schema.StringAttribute{
				Required:      true,
				Description:   "Role name (e.g., admin, developer, viewer).",
				PlanModifiers: replaceOnChange,
			},
			"project": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Project scope. Required when scope is project.",
				PlanModifiers: replaceOnChange,
			},
			"scope": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Default:       stringdefault.StaticString("project"),
				Description:   "project (default) or organization.",
				PlanModifiers: replaceOnChange,
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

func (r *RoleBindingResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{&roleBindingScopeValidator{}}
}

func (r *RoleBindingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RoleBindingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan roleBindingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.AddRoleBinding(ctx, plan.User.ValueString(), plan.Role.ValueString(), plan.Project.ValueString(), plan.Scope.ValueString()); err != nil {
		resp.Diagnostics.AddError("Create Role Binding Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.Scope.ValueString() + "/" + plan.User.ValueString() + "/" + plan.Role.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RoleBindingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state roleBindingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	binding, err := r.client.FindRoleBinding(ctx, state.User.ValueString(), state.Role.ValueString(), state.Project.ValueString(), state.Scope.ValueString())
	if err != nil {
		if akka.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Role Binding Failed", err.Error())
		return
	}

	state.User = types.StringValue(binding.User)
	state.Role = types.StringValue(binding.Role)
	state.Project = types.StringValue(binding.Project)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RoleBindingResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update Not Supported", "All role binding fields force resource recreation.")
}

func (r *RoleBindingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state roleBindingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteRoleBinding(ctx, state.User.ValueString(), state.Role.ValueString(), state.Project.ValueString(), state.Scope.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete Role Binding Failed", err.Error())
	}
}

func (r *RoleBindingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Format: <scope>/<user>/<role>  or  <project>/<user>/<role>
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid Import ID", "Expected format: <project>/<user>/<role> or organization/<user>/<role>")
		return
	}
	scope := "project"
	project := parts[0]
	if parts[0] == "organization" {
		scope = "organization"
		project = ""
	}
	state := roleBindingResourceModel{
		ID:      types.StringValue(req.ID),
		User:    types.StringValue(parts[1]),
		Role:    types.StringValue(parts[2]),
		Project: types.StringValue(project),
		Scope:   types.StringValue(scope),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

type roleBindingScopeValidator struct{}

func (v *roleBindingScopeValidator) Description(_ context.Context) string {
	return "Validates scope and project consistency."
}

func (v *roleBindingScopeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *roleBindingScopeValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config roleBindingResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scope := config.Scope.ValueString()
	project := config.Project.ValueString()

	if scope == "organization" && project != "" {
		resp.Diagnostics.AddError("Invalid Configuration", "'project' must not be set when scope is 'organization'.")
	}
}
