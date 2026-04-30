package provider

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/thinkmorestupidless/terraform-provider-akka/internal/akka"
)

var _ provider.Provider = &akkaProvider{}
var _ provider.ProviderWithFunctions = &akkaProvider{}

type akkaProvider struct {
	version string
}

type akkaProviderModel struct {
	Organization types.String `tfsdk:"organization"`
	Token        types.String `tfsdk:"token"`
	Project      types.String `tfsdk:"project"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &akkaProvider{version: version}
	}
}

func (p *akkaProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "akka"
	resp.Version = p.version
}

func (p *akkaProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages resources on the Akka platform using the akka CLI.",
		Attributes: map[string]schema.Attribute{
			"organization": schema.StringAttribute{
				Required:    true,
				Description: "Akka organization name. All CLI commands run in this org scope.",
			},
			"token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Akka API token. Falls back to the AKKA_TOKEN environment variable.",
			},
			"project": schema.StringAttribute{
				Optional:    true,
				Description: "Default project for resources that don't specify one. Falls back to AKKA_PROJECT.",
			},
		},
	}
}

func (p *akkaProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config akkaProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	binaryPath, err := exec.LookPath("akka")
	if err != nil {
		resp.Diagnostics.AddError(
			"akka CLI Not Found",
			"The provider requires the akka CLI to be installed and available in PATH. "+
				"Install it from https://doc.akka.io/install-cli.sh. Error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "akka CLI found", map[string]any{"path": binaryPath})

	token := config.Token.ValueString()
	if token == "" {
		token = os.Getenv("AKKA_TOKEN")
	}
	if token == "" {
		resp.Diagnostics.AddError(
			"Missing API Token",
			"The provider requires an Akka API token. Set the 'token' attribute or the AKKA_TOKEN environment variable.",
		)
		return
	}

	defaultProject := config.Project.ValueString()
	if defaultProject == "" {
		defaultProject = os.Getenv("AKKA_PROJECT")
	}

	client := akka.NewClient(binaryPath, token, config.Organization.ValueString(), defaultProject)
	checkAkkaCLIVersion(ctx, client, resp)
	resp.ResourceData = client
	resp.DataSourceData = client
}

// minimumTestedVersion is the oldest akka CLI version validated against this provider.
const minimumTestedVersion = "3.0.0"

func checkAkkaCLIVersion(ctx context.Context, client *akka.AkkaClient, resp *provider.ConfigureResponse) {
	out, err := client.Run(ctx, "version")
	if err != nil {
		tflog.Warn(ctx, "could not determine akka CLI version", map[string]any{"error": err.Error()})
		return
	}
	var v struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(out, &v); err != nil || v.Version == "" {
		tflog.Warn(ctx, "akka CLI version response was not parseable", map[string]any{"raw": string(out)})
		return
	}
	tflog.Debug(ctx, "akka CLI version", map[string]any{"version": v.Version})
	if v.Version < minimumTestedVersion {
		resp.Diagnostics.AddWarning(
			"akka CLI Version May Be Too Old",
			"This provider has been tested with akka CLI "+minimumTestedVersion+" and later. "+
				"Your installed version is "+v.Version+". Some operations may not work as expected. "+
				"Upgrade with: curl -sL https://doc.akka.io/install-cli.sh | bash",
		)
	}
}

func (p *akkaProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProjectResource,
		NewServiceResource,
		NewSecretResource,
		NewRouteResource,
		NewHostnameResource,
		NewRoleBindingResource,
	}
}

func (p *akkaProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewProjectDataSource,
		NewServiceDataSource,
		NewRegionsDataSource,
	}
}

func (p *akkaProvider) Functions(_ context.Context) []func() function.Function {
	return nil
}
