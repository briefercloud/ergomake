package ghlauncher

import (
	"fmt"
	"strings"

	"github.com/ergomake/ergomake/internal/transformer"
)

func createSuccessComment(env *transformer.Environment, frontendEnvLink string) string {
	return fmt.Sprintf(`Hi 👋

Here's a preview environment 🚀

%s

# Environment Summary 📑

| Container | Source | URL |
| - | - | - |
%s

Here are your environment's [logs](%s).

For questions or comments, [join Discord](https://discord.gg/daGzchUGDt).

[Click here](https://github.com/apps/ergomake) to disable Ergomake.`,
		getMainServiceUrl(env),
		getServiceTable(env),
		frontendEnvLink,
	)
}

func getMainServiceUrl(env *transformer.Environment) string {
	return getServiceUrl(env.FirstService())
}

func createFailureComment(frontendLink string, validationError *transformer.ProjectValidationError) string {
	reason := fmt.Sprintf(
		`You can see your environment build logs [here](%s). Please double-check your `+"`docker-compose.yml`"+` file is valid.`,
		frontendLink,
	)

	if validationError != nil {
		reason = validationError.Message
	}

	return fmt.Sprintf(`Hi 👋

We couldn't create a preview environment for this pull-request 😥

%s

If you need help, email us at contact@getergomake.com or join [Discord](https://discord.gg/daGzchUGDt).

[Click here](https://github.com/apps/ergomake) to disable Ergomake.`, reason)
}

func createLimitedComment() string {
	return `Hi there 👋

You’ve just reached your simultaneous environments limit.

Please talk to us at contact@ergomake.dev to bump your limits.

Alternatively, you can close a PR with an existing environment, and reopen this one to get a preview.

Thanks for using Ergomake!

[Click here](https://github.com/apps/ergomake) to disable Ergomake.`
}

func getServiceTable(env *transformer.Environment) string {
	rows := make([]string, len(env.Services))
	for serviceName, serviceConfig := range env.Services {
		rows[serviceConfig.Index] = fmt.Sprintf("| %s | %s | %s |", serviceName, getSource(serviceConfig), getServiceUrl(serviceConfig))
	}
	return strings.Join(rows, "\n")
}

func getServiceUrl(svc transformer.EnvironmentService) string {
	if svc.Url == "" {
		return "[not exposed - internal service]"
	}

	return fmt.Sprintf("https://%s", svc.Url)
}

func getSource(svc transformer.EnvironmentService) string {
	if svc.Build != "" {
		return "Dockerfile"
	}

	return svc.Image
}
