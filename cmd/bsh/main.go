package main

import (
	"os"

	fmt "github.com/jhunt/go-ansi"
	cli "github.com/jhunt/go-cli"

	"github.com/jhunt/bsh/bosh"
)

var Version string

const (
	OopsBadOptions int = iota
	OopsNotImplemented
	OopsBadConfiguration
	OopsCommunicationFailed
	OopsSaveConfigFailed
	OopsJSONFailed
	OopsTaskFailed
	OopsCancelled
	OopsDownloadFailed
)

type Opt struct {
	Help    bool `cli:"-h, --help"`
	Version bool `cli:"-v"`

	URL      string `cli:"--director, --url"`
	Username string `cli:"-u, --username"`
	Password string `cli:"-p, --password"`
	CaCert   string `cli:"--ca-cert"`
	Insecure bool   `cli:"-k, --insecure, --no-insecure"`

	Config     string `cli:"-c, --config"`
	BOSHTarget string `cli:"-t, --target"`

	AsJSON bool `cli:"--json"`
	Batch  bool `cli:"-y, --yes"`

	Deployment string `cli:"-d, --deployment"`

	Deploy struct {
		Recreate bool `cli:"-R, --recreate"`
		Redact   bool `cli:"--redact"`
	} `cli:"deploy"`

	Diff struct {
		Redact bool `cli:"--redact"`
	} `cli:"diff"`

	Task struct {
		Event  bool `cli:"--event"`
		Debug  bool `cli:"--debug"`
		Result bool `cli:"--result"`
		Raw    bool `cli:"--raw"`
		CPI    bool `cli:"--cpi"`
	} `cli:"task"`

	Check struct{} `cli:"check"`

	Tasks struct {
		All       bool     `cli:"-a, --all"`
		States    []string `cli:"-s, --state"`
		ContextID string   `cli:"-C, --context, --context-id"`
		Limit     int      `cli:"-l, --limit"`
	} `cli:"tasks"`

	Cleanup struct {
		All bool `cli:"-a, --all"`
	} `cli:"cleanup"`

	Curl struct{} `cli:"curl"`

	Locks struct{} `cli:"locks"`

	Deployments struct{} `cli:"deployments"`

	Releases struct {
		Jobs bool `cli:"--jobs"`
	} `cli:"releases"`

	Stemcells struct{} `cli:"stemcells"`
	Errands   struct{} `cli:"errands"`
	Login     struct{} `cli:"login"`
	Logout    struct{} `cli:"logout"`
	Status    struct{} `cli:"status"`

	VMs struct {
		Vitals    bool `cli:"--vitals"`
		Details   bool `cli:"--details"`
		Processes bool `cli:"--ps"`
		DNS       bool `cli:"--dns"`
	} `cli:"vms, instances"`

	Targets struct{} `cli:"targets"`
	Target  struct{} `cli:"target"`

	Delete struct {
		Force      bool     `cli:"-f, --force"`
		Release    struct{} `cli:"release"`
		Stemcell   struct{} `cli:"stemcell"`
		Deployment struct{} `cli:"deployment"`
		VM         struct{} `cli:"vm"`
	} `cli:"delete"`

	Upload struct {
		Fix          bool   `cli:"--fix"`
		SkipIfExists bool   `cli:"--skip-if-exists"`
		SHA1         string `cli:"--sha1"`
		Name         string `cli:"--name"`
		Version      string `cli:"--version"`

		Release struct {
			Rebase bool `cli:"--rebase"`
		} `cli:"release"`

		Stemcell struct{} `cli:"stemcell"`
	} `cli:"upload"`

	Download struct {
		Output string `cli:"-o, --output"`
		Force  bool   `cli:"-f, --force"`

		Manifest struct {
		} `cli:"manifest"`

		CloudConfig struct {
		} `cli:"cloud-config"`

		RuntimeConfig struct {
		} `cli:"runtime-config"`
	} `cli:"download"`

	Ignore struct {
		Instance struct{} `cli:"instance"`
	} `cli:"ignore"`

	Unignore struct {
		Instance struct{} `cli:"instance"`
	} `cli:"unignore"`

	Inspect struct {
		Release struct {
		} `cli:"release"`
	} `cli:"inspect"`

	Start struct {
		Soft        bool `cli:"--soft"`
		Hard        bool `cli:"--hard"`
		Force       bool `cli:"--force"`
		Fix         bool `cli:"--fix"`
		DryRun      bool `cli:"--dry-run"`
		MaxInFlight int  `cli:"--max-in-flight"`
		Canaries    int  `cli:"--canaries"`
		SkipDrain   bool `cli:"--ski-drain"`
	} `cli:"start"`

	Stop struct {
		Soft        bool `cli:"--soft"`
		Hard        bool `cli:"--hard"`
		Force       bool `cli:"--force"`
		DryRun      bool `cli:"--dry-run"`
		MaxInFlight int  `cli:"--max-in-flight"`
		Canaries    int  `cli:"--canaries"`
		SkipDrain   bool `cli:"--ski-drain"`
	} `cli:"stop"`

	Restart struct {
		Force       bool `cli:"--force"`
		DryRun      bool `cli:"--dry-run"`
		MaxInFlight int  `cli:"--max-in-flight"`
		Canaries    int  `cli:"--canaries"`
		SkipDrain   bool `cli:"--ski-drain"`
	} `cli:"restart"`

	Recreate struct {
		Force       bool `cli:"--force"`
		Fix         bool `cli:"--fix"`
		DryRun      bool `cli:"--dry-run"`
		MaxInFlight int  `cli:"--max-in-flight"`
		Canaries    int  `cli:"--canaries"`
		SkipDrain   bool `cli:"--ski-drain"`
	} `cli:"recreate"`
}

func main() {
	var opt Opt
	opt.Config = fmt.Sprintf("%s/%s", os.Getenv("HOME"), ".boshrc")

	opt.Tasks.Limit = 30

	/* make sure ~/.boshrc exists... */
	if err := bosh.DefaultConfig(opt.Config); err != nil {
		fmt.Fprintf(os.Stderr, "%s: @Y{%s}\n", opt.Config, err)
	}

	command, args, err := cli.Parse(&opt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "@R{!!! %s}\n", err)
		os.Exit(OopsBadOptions)
	}

	if command == "" && len(args) > 0 && args[0] == "version" {
		opt.Version = true
	}

	if opt.Version {
		if Version == "" {
			fmt.Printf("bsh @*{development version} ... @C{¯\\_(ツ)_/¯}\n")
		} else {
			fmt.Printf("bsh %s\n", Version)
		}
		os.Exit(0)
	}

	if command == "" {
		fmt.Fprintf(os.Stderr, "@R{a command is required...}\n")
		os.Exit(OopsBadOptions)
	}

	known := map[string]func(Opt, string, []string){
		"check":             runCheck,
		"cleanup":           runCleanup,
		"curl":              runCurl,
		"deploy":            runDeploy,
		"deployments":       runDeployments,
		"diff":              runDiff,
		"locks":             runLocks,
		"login":             runLogin,
		"logout":            runLogout,
		"releases":          runReleases,
		"status":            runStatus,
		"stemcells":         runStemcells,
		"target":            runTarget,
		"targets":           runTargets,
		"task":              runTask,
		"tasks":             runTasks,
		"vms":               runVMs,
		"errands":           runErrands,
		"ignore instance":   runIgnoreInstance,
		"unignore instance": runUnignoreInstance,

		"delete release":    runDeleteRelease,
		"delete stemcell":   runDeleteStemcell,
		"delete deployment": runDeleteDeployment,
		"delete vm":         runDeleteVM,

		"inspect release": runInspectRelease,

		"upload release":  runUploadRelease,
		"upload stemcell": runUploadStemcell,

		"download manifest":       runDownloadManifest,
		"download cloud-config":   runDownloadCloudConfig,
		"download runtime-config": runDownloadRuntimeConfig,

		"start":    runStart,
		"stop":     runStop,
		"restart":  runRestart,
		"recreate": runRecreate,
	}

	if fn, ok := known[command]; ok {
		fn(opt, command, args)
		os.Exit(0)
	}
	fmt.Fprintf(os.Stderr, "%s - @*{not yet implemented...}\n", command)
	os.Exit(OopsNotImplemented)
}
