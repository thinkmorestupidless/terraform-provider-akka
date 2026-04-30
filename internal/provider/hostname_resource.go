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

var _ resource.Resource = &HostnameResource{}
var _ resource.ResourceWithImportState = &HostnameResource{}

type HostnameResource struct {
	client *akka.AkkaClient
}

type hostnameResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Hostname     types.String `tfsdk:"hostname"`
	Project      types.String `tfsdk:"project"`
	TLSSecret    types.String `tfsdk:"tls_secret"`
	Status       types.String `tfsdk:"status"`
	Organization types.String `tfsdk:"organization"`
}

func NewHostnameResource() resource.Resource {
	return &HostnameResource{}
}

func (r *HostnameResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hostname"
}

func (r *HostnameResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a custom domain name registered on an Akka project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hostname": schema.StringAttribute{
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
			"tls_secret": schema.StringAttribute{
				Optional:    true,
				Description: "Name of an akka_secret of type tls to use for this hostname.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Verification status: Pending, Verified, Failed.",
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
		},
	}
}

func (r *HostnameResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *HostnameResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan hostnameResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	h, err := r.client.CreateHostname(ctx, plan.Hostname.ValueString(), plan.Project.ValueString(), plan.TLSSecret.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Create Hostname Failed", err.Error())
		return
	}
	mapHostnameToState(h, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *HostnameResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state hostnameResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	h, err := r.client.GetHostname(ctx, state.Hostname.ValueString(), state.Project.ValueString())
	if err != nil {
		if akka.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Hostname Failed", err.Error())
		return
	}
	mapHostnameToState(h, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *HostnameResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan hostnameResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Only tls_secret can change in-place; hostname requires replace.
	h, err := r.client.GetHostname(ctx, plan.Hostname.ValueString(), plan.Project.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Hostname During Update Failed", err.Error())
		return
	}
	h.TLSSecret = plan.TLSSecret.ValueString()
	mapHostnameToState(h, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *HostnameResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state hostnameResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteHostname(ctx, state.Hostname.ValueString(), state.Project.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete Hostname Failed", err.Error())
	}
}

func (r *HostnameResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid Import ID", "Expected format: <project>/<hostname>")
		return
	}
	h, err := r.client.GetHostname(ctx, parts[1], parts[0])
	if err != nil {
		resp.Diagnostics.AddError("Import Hostname Failed", err.Error())
		return
	}
	var state hostnameResourceModel
	mapHostnameToState(h, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func mapHostnameToState(h *akka.HostnameModel, m *hostnameResourceModel) {
	m.ID = types.StringValue(h.Project + "/" + h.Hostname)
	m.Hostname = types.StringValue(h.Hostname)
	m.Project = types.StringValue(h.Project)
	m.TLSSecret = types.StringValue(h.TLSSecret)
	m.Status = types.StringValue(h.Status)
}
