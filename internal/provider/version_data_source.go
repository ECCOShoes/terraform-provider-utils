// Copyright (c) ECCO A/S
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource = &VersionDataSource{}
)

func NewVersionDataSource() datasource.DataSource {
	return &VersionDataSource{}
}

type VersionDataSource struct{}

// major.minor only: rev_number/build_number are appended separately.
var baseVersionPattern = regexp.MustCompile(`^\d+\.\d+$`)

var debianDatePattern = regexp.MustCompile(`^\d{8}$`)

type versionDataSourceModel struct {
	BaseVersion    types.String `tfsdk:"base_version"`
	RevNumber      types.Int64  `tfsdk:"rev_number"`
	BuildNumber    types.Int64  `tfsdk:"build_number"`
	GitSha         types.String `tfsdk:"git_sha"`
	ReleaseChannel types.String `tfsdk:"release_channel"`
	DebianRevision types.String `tfsdk:"debian_revision"`
	DebianDate     types.String `tfsdk:"debian_date"`
	Arch           types.String `tfsdk:"arch"`
	Dotnet         types.Object `tfsdk:"dotnet"`
	Debian         types.Object `tfsdk:"debian"`
	Nuget          types.Object `tfsdk:"nuget"`
	Npm            types.Object `tfsdk:"npm"`
}

type dotnetVersionModel struct {
	AppVersion  types.String `tfsdk:"app_version"`
	InfoVersion types.String `tfsdk:"info_version"`
}

type debianVersionModel struct {
	Version types.String `tfsdk:"version"`
	Arch    types.String `tfsdk:"arch"`
}

type semverVersionModel struct {
	Version types.String `tfsdk:"version"`
}

var dotnetAttrTypes = map[string]attr.Type{
	"app_version":  types.StringType,
	"info_version": types.StringType,
}

var debianAttrTypes = map[string]attr.Type{
	"version": types.StringType,
	"arch":    types.StringType,
}

var semverAttrTypes = map[string]attr.Type{
	"version": types.StringType,
}

func (d *VersionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_version"
}

func (d *VersionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Computes ecosystem-specific version strings (dotnet, debian, nuget, npm) from a common set of build inputs (base version, revision/build numbers, git commit sha, and release channel).",
		Attributes: map[string]schema.Attribute{
			"base_version": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Base version in `major.minor` format (e.g. `\"1.2\"`). The revision number is appended as the patch component for all ecosystems, and the build number is appended as a fourth component for dotnet.",
			},
			"rev_number": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Revision number (e.g. git commit count), used as the patch component of all computed versions.",
			},
			"build_number": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "CI build number. Used as the fourth numeric component for dotnet, and as the prerelease build identifier for nuget/npm.",
			},
			"git_sha": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Short git commit sha. Omitted from all output version strings if not set.",
			},
			"release_channel": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "One of `\"release\"`, `\"beta\"`, or `\"alpha\"`. Defaults to `\"alpha\"`. `\"release\"` produces plain numeric versions; `\"beta\"`/`\"alpha\"` add a prerelease suffix.",
			},
			"debian_revision": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Debian package revision (the trailing `-N` in the version string). Defaults to `\"1\"`.",
			},
			"debian_date": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Date used in the debian `~git<date>` snapshot segment, in `YYYYMMDD` format. Defaults to the current UTC date; override for reproducible builds.",
			},
			"arch": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Target architecture, echoed back as `debian.arch` for building package filenames. Not embedded into any version string.",
			},
			"dotnet": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Version strings for .NET assembly/package versioning.",
				Attributes: map[string]schema.Attribute{
					"app_version": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Purely numeric 4-part version (`major.minor.rev.build`), suitable for `AssemblyVersion`/`FileVersion`.",
					},
					"info_version": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "`major.minor.rev` on the release channel, or `major.minor-<channel>` otherwise; `-<sha>` is appended whenever `git_sha` is set, regardless of channel.",
					},
				},
			},
			"debian": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Version string for a Debian/APT package.",
				Attributes: map[string]schema.Attribute{
					"version": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "e.g. `1.2.3-1` on the release channel, or `1.2.3~alpha~git20240101.abc1234-1` otherwise. If `git_sha` is set on the release channel, it is appended as `+git<sha>`.",
					},
					"arch": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Echoes the `arch` input, null if not set.",
					},
				},
			},
			"nuget": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Version string for a NuGet package (SemVer 2.0.0).",
				Attributes: map[string]schema.Attribute{
					"version": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "e.g. `1.2.3` on the release channel (`1.2.3+abc1234` if `git_sha` is set), or `1.2.3-alpha.4+abc1234` otherwise.",
					},
				},
			},
			"npm": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Version string for an npm package (SemVer 2.0.0).",
				Attributes: map[string]schema.Attribute{
					"version": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "e.g. `1.2.3` on the release channel (`1.2.3+abc1234` if `git_sha` is set), or `1.2.3-alpha.4+abc1234` otherwise.",
					},
				},
			},
		},
	}
}

func (d *VersionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data versionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	baseVersion := data.BaseVersion.ValueString()
	if !baseVersionPattern.MatchString(baseVersion) {
		resp.Diagnostics.AddAttributeError(path.Root("base_version"), "Invalid base_version",
			fmt.Sprintf("base_version must be in \"major.minor\" format (e.g. \"1.2\"), got %q", baseVersion))
	}

	revNumber := data.RevNumber.ValueInt64()
	if revNumber < 0 {
		resp.Diagnostics.AddAttributeError(path.Root("rev_number"), "Invalid rev_number", "rev_number must not be negative")
	}

	buildNumber := data.BuildNumber.ValueInt64()
	if buildNumber < 0 {
		resp.Diagnostics.AddAttributeError(path.Root("build_number"), "Invalid build_number", "build_number must not be negative")
	}

	releaseChannel := "alpha"
	if !data.ReleaseChannel.IsNull() && !data.ReleaseChannel.IsUnknown() {
		releaseChannel = data.ReleaseChannel.ValueString()
	}
	switch releaseChannel {
	case "release", "beta", "alpha":
	default:
		resp.Diagnostics.AddAttributeError(path.Root("release_channel"), "Invalid release_channel",
			fmt.Sprintf("release_channel must be one of \"release\", \"beta\", or \"alpha\", got %q", releaseChannel))
	}

	debianDate := time.Now().UTC().Format("20060102")
	if !data.DebianDate.IsNull() && !data.DebianDate.IsUnknown() {
		debianDate = data.DebianDate.ValueString()
		if !debianDatePattern.MatchString(debianDate) {
			resp.Diagnostics.AddAttributeError(path.Root("debian_date"), "Invalid debian_date",
				fmt.Sprintf("debian_date must be in \"YYYYMMDD\" format, got %q", debianDate))
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	gitSha := ""
	if !data.GitSha.IsNull() && !data.GitSha.IsUnknown() {
		gitSha = data.GitSha.ValueString()
	}

	debianRevision := "1"
	if !data.DebianRevision.IsNull() && !data.DebianRevision.IsUnknown() {
		debianRevision = data.DebianRevision.ValueString()
	}

	isRelease := releaseChannel == "release"
	// major.minor.rev, the common numeric core shared by nuget/npm/debian.
	semverBase := fmt.Sprintf("%s.%d", baseVersion, revNumber)

	dotnetAppVersion := fmt.Sprintf("%s.%d.%d", baseVersion, revNumber, buildNumber)
	dotnetInfoVersion := semverBase
	if !isRelease {
		dotnetInfoVersion = fmt.Sprintf("%s-%s", baseVersion, releaseChannel)
	}
	if gitSha != "" {
		dotnetInfoVersion += "-" + gitSha
	}

	semverVersion := semverBase
	if !isRelease {
		semverVersion = fmt.Sprintf("%s-%s.%d", semverBase, releaseChannel, buildNumber)
	}
	if gitSha != "" {
		semverVersion += "+" + gitSha
	}

	debianVersion := semverBase
	if !isRelease {
		debianVersion = fmt.Sprintf("%s~%s~git%s", semverBase, releaseChannel, debianDate)
	}
	if gitSha != "" {
		if isRelease {
			debianVersion += "+git" + gitSha
		} else {
			debianVersion += "." + gitSha
		}
	}
	debianVersion = fmt.Sprintf("%s-%s", debianVersion, debianRevision)

	dotnetObj, diags := types.ObjectValueFrom(ctx, dotnetAttrTypes, dotnetVersionModel{
		AppVersion:  types.StringValue(dotnetAppVersion),
		InfoVersion: types.StringValue(dotnetInfoVersion),
	})
	resp.Diagnostics.Append(diags...)

	debianObj, diags := types.ObjectValueFrom(ctx, debianAttrTypes, debianVersionModel{
		Version: types.StringValue(debianVersion),
		Arch:    data.Arch,
	})
	resp.Diagnostics.Append(diags...)

	nugetObj, diags := types.ObjectValueFrom(ctx, semverAttrTypes, semverVersionModel{
		Version: types.StringValue(semverVersion),
	})
	resp.Diagnostics.Append(diags...)

	npmObj, diags := types.ObjectValueFrom(ctx, semverAttrTypes, semverVersionModel{
		Version: types.StringValue(semverVersion),
	})
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.ReleaseChannel = types.StringValue(releaseChannel)
	data.DebianRevision = types.StringValue(debianRevision)
	data.DebianDate = types.StringValue(debianDate)
	data.Dotnet = dotnetObj
	data.Debian = debianObj
	data.Nuget = nugetObj
	data.Npm = npmObj

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
