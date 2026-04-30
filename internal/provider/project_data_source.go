package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thinkmorestupidless/terraform-provider-akka/internal/akka"
)

var _ datasource.DataSource = &ProjectDataSource{}

type ProjectDataSource struct {
	client *akka.AkkaClient
}

type projectDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Organization types.String `tfsdk:"organization"`
	Description  types.String `tfsdk:"description"`
	Region       types.String `tfsdk:"region"`
}

func NewProjectDataSource() datasource.DataSource {
	return &ProjectDataSource{}
}

func (d *ProjectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *ProjectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing Akka project without managing its lifecycle.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Project name to look up.",
			},
			"organization": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Organization override.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Project description.",
			},
			"region": schema.StringAttribute{
				Computed:    true,
				Description: "Primary deployment region.",
			},
		},
	}
}

func (d *ProjectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state projectDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := d.client.GetProject(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Project Data Source Failed", err.Error())
		return
	}

	state.ID = types.StringValue(project.ID)
	state.Name = types.StringValue(project.Name)
	state.Description = types.StringValue(project.Description)
	state.Region = types.StringValue(project.Region)
	state.Organization = types.StringValue(project.Organization)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
