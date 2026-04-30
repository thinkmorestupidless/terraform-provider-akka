package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thinkmorestupidless/terraform-provider-akka/internal/akka"
)

var _ resource.Resource = &RouteResource{}
var _ resource.ResourceWithImportState = &RouteResource{}

type RouteResource struct {
	client *akka.AkkaClient
}

type routeResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Project      types.String `tfsdk:"project"`
	Hostname     types.String `tfsdk:"hostname"`
	Paths        types.Map    `tfsdk:"paths"`
	Organization types.String `tfsdk:"organization"`
}

func NewRouteResource() resource.Resource {
	return &RouteResource{}
}

func (r *RouteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_route"
}

func (r *RouteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Akka traffic route mapping a hostname and path prefixes to services.",
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
			"hostname": schema.StringAttribute{Required: true},
			"paths": schema.MapAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Map of URL path prefix to service name.",
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

func (r *RouteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RouteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan routeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	route, err := r.client.CreateRoute(ctx, plan.Name.ValueString(), plan.Project.ValueString(), plan.Hostname.ValueString(), toStringMap(ctx, plan.Paths))
	if err != nil {
		resp.Diagnostics.AddError("Create Route Failed", err.Error())
		return
	}

	mapRouteToState(route, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RouteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state routeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	route, err := r.client.GetRoute(ctx, state.Name.ValueString(), state.Project.ValueString())
	if err != nil {
		if akka.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Route Failed", err.Error())
		return
	}
	mapRouteToState(route, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RouteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan routeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	route, err := r.client.UpdateRoute(ctx, plan.Name.ValueString(), plan.Project.ValueString(), plan.Hostname.ValueString(), toStringMap(ctx, plan.Paths))
	if err != nil {
		resp.Diagnostics.AddError("Update Route Failed", err.Error())
		return
	}
	mapRouteToState(route, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RouteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state routeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteRoute(ctx, state.Name.ValueString(), state.Project.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete Route Failed", err.Error())
	}
}

func (r *RouteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid Import ID", "Expected format: <project>/<name>")
		return
	}
	route, err := r.client.GetRoute(ctx, parts[1], parts[0])
	if err != nil {
		resp.Diagnostics.AddError("Import Route Failed", err.Error())
		return
	}
	var state routeResourceModel
	mapRouteToState(route, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func mapRouteToState(route *akka.RouteModel, m *routeResourceModel) {
	m.ID = types.StringValue(route.Project + "/" + route.Name)
	m.Name = types.StringValue(route.Name)
	m.Project = types.StringValue(route.Project)
	m.Hostname = types.StringValue(route.Host)

	pathElems := make(map[string]attr.Value, len(route.Paths))
	for k, v := range route.Paths {
		pathElems[k] = types.StringValue(v)
	}
	m.Paths, _ = types.MapValue(types.StringType, pathElems)
}
