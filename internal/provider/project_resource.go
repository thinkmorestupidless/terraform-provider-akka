package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thinkmorestupidless/terraform-provider-akka/internal/akka"
)

var _ resource.Resource = &ProjectResource{}
var _ resource.ResourceWithImportState = &ProjectResource{}

type ProjectResource struct {
	client *akka.AkkaClient
}

type projectResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Region       types.String `tfsdk:"region"`
	Organization types.String `tfsdk:"organization"`
}

func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

func (r *ProjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *ProjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Akka project within the configured organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal Akka project ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Project name. Unique within the organization. Changing forces recreation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Human-readable project description.",
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Primary deployment region.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Organization override. Defaults to provider-level organization.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ProjectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.client.CreateProject(ctx, plan.Name.ValueString(), plan.Description.ValueString(), plan.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Create Project Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(project.ID)
	plan.Name = types.StringValue(project.Name)
	plan.Description = types.StringValue(project.Description)
	plan.Region = types.StringValue(project.Region)
	plan.Organization = types.StringValue(project.Organization)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.client.GetProject(ctx, state.Name.ValueString())
	if err != nil {
		if akka.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Project Failed", err.Error())
		return
	}

	state.ID = types.StringValue(project.ID)
	state.Name = types.StringValue(project.Name)
	state.Description = types.StringValue(project.Description)
	state.Region = types.StringValue(project.Region)
	state.Organization = types.StringValue(project.Organization)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.client.UpdateProject(ctx, plan.Name.ValueString(), plan.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Update Project Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(project.ID)
	plan.Description = types.StringValue(project.Description)
	plan.Region = types.StringValue(project.Region)
	plan.Organization = types.StringValue(project.Organization)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteProject(ctx, state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete Project Failed", err.Error())
	}
}

func (r *ProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by project name
	project, err := r.client.GetProject(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Project Failed", err.Error())
		return
	}
	state := projectResourceModel{
		ID:           types.StringValue(project.ID),
		Name:         types.StringValue(project.Name),
		Description:  types.StringValue(project.Description),
		Region:       types.StringValue(project.Region),
		Organization: types.StringValue(project.Organization),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
