package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thinkmorestupidless/terraform-provider-akka/internal/akka"
)

var _ datasource.DataSource = &ServiceDataSource{}

type ServiceDataSource struct {
	client *akka.AkkaClient
}

type serviceDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Project      types.String `tfsdk:"project"`
	Image        types.String `tfsdk:"image"`
	Status       types.String `tfsdk:"status"`
	Hostname     types.String `tfsdk:"hostname"`
	Exposed      types.Bool   `tfsdk:"exposed"`
	Organization types.String `tfsdk:"organization"`
}

func NewServiceDataSource() datasource.DataSource {
	return &ServiceDataSource{}
}

func (d *ServiceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service"
}

func (d *ServiceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing Akka service without managing its lifecycle.",
		Attributes: map[string]schema.Attribute{
			"id":           schema.StringAttribute{Computed: true},
			"name":         schema.StringAttribute{Required: true},
			"project":      schema.StringAttribute{Optional: true, Computed: true},
			"image":        schema.StringAttribute{Computed: true},
			"status":       schema.StringAttribute{Computed: true},
			"hostname":     schema.StringAttribute{Computed: true},
			"exposed":      schema.BoolAttribute{Computed: true},
			"organization": schema.StringAttribute{Optional: true, Computed: true},
		},
	}
}

func (d *ServiceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*akka.AkkaClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("expected *akka.AkkaClient, got %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *ServiceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state serviceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	svc, err := d.client.GetService(ctx, state.Name.ValueString(), state.Project.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Service Data Source Failed", err.Error())
		return
	}

	state.ID = types.StringValue(svc.ID)
	state.Name = types.StringValue(svc.Name)
	state.Project = types.StringValue(svc.Project)
	state.Image = types.StringValue(svc.Image)
	state.Status = types.StringValue(svc.Status)
	state.Hostname = types.StringValue(svc.Hostname)
	state.Exposed = types.BoolValue(svc.Exposed)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
