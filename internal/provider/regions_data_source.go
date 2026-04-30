package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thinkmorestupidless/terraform-provider-akka/internal/akka"
)

var _ datasource.DataSource = &RegionsDataSource{}

type RegionsDataSource struct {
	client *akka.AkkaClient
}

type regionsDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Organization types.String `tfsdk:"organization"`
	Regions      types.List   `tfsdk:"regions"`
}

var regionObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"name":         types.StringType,
		"display_name": types.StringType,
		"provider":     types.StringType,
	},
}

func NewRegionsDataSource() datasource.DataSource {
	return &RegionsDataSource{}
}

func (d *RegionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_regions"
}

func (d *RegionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all deployment regions available in the organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"organization": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Organization override.",
			},
			"regions": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of available deployment regions.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":         schema.StringAttribute{Computed: true},
						"display_name": schema.StringAttribute{Computed: true},
						"provider":     schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *RegionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RegionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state regionsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	regions, err := d.client.ListRegions(ctx)
	if err != nil {
		resp.Diagnostics.AddError("List Regions Failed", err.Error())
		return
	}

	regionValues := make([]attr.Value, len(regions))
	for i, r := range regions {
		obj, diags := types.ObjectValue(regionObjectType.AttrTypes, map[string]attr.Value{
			"name":         types.StringValue(r.Name),
			"display_name": types.StringValue(r.DisplayName),
			"provider":     types.StringValue(r.Provider),
		})
		resp.Diagnostics.Append(diags...)
		regionValues[i] = obj
	}

	regionList, diags := types.ListValue(regionObjectType, regionValues)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue("regions")
	state.Regions = regionList
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
