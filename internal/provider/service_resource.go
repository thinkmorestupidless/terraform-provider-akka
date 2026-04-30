package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thinkmorestupidless/terraform-provider-akka/internal/akka"
)

var _ resource.Resource = &ServiceResource{}
var _ resource.ResourceWithImportState = &ServiceResource{}

type ServiceResource struct {
	client *akka.AkkaClient
}

type serviceResourceModel struct {
	ID           types.String   `tfsdk:"id"`
	Name         types.String   `tfsdk:"name"`
	Project      types.String   `tfsdk:"project"`
	Image        types.String   `tfsdk:"image"`
	Region       types.String   `tfsdk:"region"`
	Replicas     types.Int64    `tfsdk:"replicas"`
	Env          types.Map      `tfsdk:"env"`
	Exposed      types.Bool     `tfsdk:"exposed"`
	Paused       types.Bool     `tfsdk:"paused"`
	Status       types.String   `tfsdk:"status"`
	Hostname     types.String   `tfsdk:"hostname"`
	Organization types.String   `tfsdk:"organization"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
}

func NewServiceResource() resource.Resource {
	return &ServiceResource{}
}

func (r *ServiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service"
}

func (r *ServiceResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a deployed Akka service within a project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"image":    schema.StringAttribute{Required: true},
			"region":   schema.StringAttribute{Optional: true, Computed: true},
			"replicas": schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(1)},
			"env": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"exposed": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"paused":  schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"status": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hostname": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *ServiceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serviceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	env := toStringMap(ctx, plan.Env)
	svc, err := r.client.DeployService(ctx,
		plan.Name.ValueString(),
		plan.Project.ValueString(),
		plan.Image.ValueString(),
		plan.Region.ValueString(),
		int(plan.Replicas.ValueInt64()),
		env,
	)
	if err != nil {
		resp.Diagnostics.AddError("Deploy Service Failed", err.Error())
		return
	}

	if err := r.client.WaitForReady(ctx, svc.Name, svc.Project, 10*time.Second); err != nil {
		resp.Diagnostics.AddError("Service Did Not Become Ready", err.Error())
		return
	}

	if plan.Exposed.ValueBool() {
		if err := r.client.ExposeService(ctx, svc.Name, svc.Project); err != nil {
			resp.Diagnostics.AddError("Expose Service Failed", err.Error())
			return
		}
	}

	svc, err = r.client.GetService(ctx, svc.Name, svc.Project)
	if err != nil {
		resp.Diagnostics.AddError("Read Service After Create Failed", err.Error())
		return
	}

	mapServiceToState(ctx, svc, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serviceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	svc, err := r.client.GetService(ctx, state.Name.ValueString(), state.Project.ValueString())
	if err != nil {
		if akka.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Service Failed", err.Error())
		return
	}

	mapServiceToState(ctx, svc, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state serviceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	// Handle pause/resume toggle
	if plan.Paused.ValueBool() != state.Paused.ValueBool() {
		if plan.Paused.ValueBool() {
			if err := r.client.PauseService(ctx, plan.Name.ValueString(), plan.Project.ValueString()); err != nil {
				resp.Diagnostics.AddError("Pause Service Failed", err.Error())
				return
			}
		} else {
			if err := r.client.ResumeService(ctx, plan.Name.ValueString(), plan.Project.ValueString()); err != nil {
				resp.Diagnostics.AddError("Resume Service Failed", err.Error())
				return
			}
		}
	}

	// Handle expose/unexpose toggle
	if plan.Exposed.ValueBool() != state.Exposed.ValueBool() {
		if plan.Exposed.ValueBool() {
			if err := r.client.ExposeService(ctx, plan.Name.ValueString(), plan.Project.ValueString()); err != nil {
				resp.Diagnostics.AddError("Expose Service Failed", err.Error())
				return
			}
		} else {
			if err := r.client.UnexposeService(ctx, plan.Name.ValueString(), plan.Project.ValueString()); err != nil {
				resp.Diagnostics.AddError("Unexpose Service Failed", err.Error())
				return
			}
		}
	}

	// Handle image/env/replicas change — redeploy
	if !plan.Image.Equal(state.Image) || !plan.Env.Equal(state.Env) || !plan.Replicas.Equal(state.Replicas) {
		env := toStringMap(ctx, plan.Env)
		if _, err := r.client.DeployService(ctx,
			plan.Name.ValueString(),
			plan.Project.ValueString(),
			plan.Image.ValueString(),
			plan.Region.ValueString(),
			int(plan.Replicas.ValueInt64()),
			env,
		); err != nil {
			resp.Diagnostics.AddError("Redeploy Service Failed", err.Error())
			return
		}
		if !plan.Paused.ValueBool() {
			if err := r.client.WaitForReady(ctx, plan.Name.ValueString(), plan.Project.ValueString(), 10*time.Second); err != nil {
				resp.Diagnostics.AddError("Service Did Not Become Ready After Update", err.Error())
				return
			}
		}
	}

	svc, err := r.client.GetService(ctx, plan.Name.ValueString(), plan.Project.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Service After Update Failed", err.Error())
		return
	}
	mapServiceToState(ctx, svc, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteService(ctx, state.Name.ValueString(), state.Project.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete Service Failed", err.Error())
	}
}

func (r *ServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid Import ID", "Expected format: <project>/<name>")
		return
	}
	svc, err := r.client.GetService(ctx, parts[1], parts[0])
	if err != nil {
		resp.Diagnostics.AddError("Import Service Failed", err.Error())
		return
	}
	var state serviceResourceModel
	mapServiceToState(ctx, svc, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func mapServiceToState(ctx context.Context, svc *akka.ServiceModel, m *serviceResourceModel) {
	m.ID = types.StringValue(svc.ID)
	m.Name = types.StringValue(svc.Name)
	m.Project = types.StringValue(svc.Project)
	m.Image = types.StringValue(svc.Image)
	m.Region = types.StringValue(svc.Region)
	m.Replicas = types.Int64Value(int64(svc.Replicas))
	m.Status = types.StringValue(svc.Status)
	m.Hostname = types.StringValue(svc.Hostname)
	m.Exposed = types.BoolValue(svc.Exposed)
	m.Paused = types.BoolValue(svc.Paused)

	envElems := make(map[string]attr.Value, len(svc.Env))
	for k, v := range svc.Env {
		envElems[k] = types.StringValue(v)
	}
	m.Env, _ = types.MapValue(types.StringType, envElems)
}

func toStringMap(ctx context.Context, m types.Map) map[string]string {
	result := make(map[string]string)
	if m.IsNull() || m.IsUnknown() {
		return result
	}
	elements := m.Elements()
	for k, v := range elements {
		if sv, ok := v.(types.String); ok {
			result[k] = sv.ValueString()
		}
	}
	return result
}
