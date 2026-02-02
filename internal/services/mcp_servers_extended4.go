package services

// ==================== MCP SERVERS - EXTENDED PART 4 ====================
// Business, CRM, HR, DevOps, Security, IoT, Government

// RegisterBusinessMCP registra servidores de Business/CRM
func (cc *ClaudeCode) RegisterBusinessMCP() {
	// Salesforce
	cc.mcpServers["salesforce"] = &MCPServer{
		Name:        "Salesforce",
		Description: "CRM Salesforce",
		Command:     "npx",
		Args:        []string{"-y", "salesforce-mcp-server"},
		Env:         map[string]string{"SF_CLIENT_ID": "", "SF_CLIENT_SECRET": "", "SF_INSTANCE_URL": ""},
		IsEnabled:   false,
	}

	// HubSpot
	cc.mcpServers["hubspot"] = &MCPServer{
		Name:        "HubSpot",
		Description: "CRM e Marketing HubSpot",
		Command:     "npx",
		Args:        []string{"-y", "hubspot-mcp-server"},
		Env:         map[string]string{"HUBSPOT_API_KEY": ""},
		IsEnabled:   false,
	}

	// Pipedrive
	cc.mcpServers["pipedrive"] = &MCPServer{
		Name:        "Pipedrive",
		Description: "CRM de vendas Pipedrive",
		Command:     "npx",
		Args:        []string{"-y", "pipedrive-mcp-server"},
		Env:         map[string]string{"PIPEDRIVE_API_TOKEN": ""},
		IsEnabled:   false,
	}

	// Zoho CRM
	cc.mcpServers["zoho-crm"] = &MCPServer{
		Name:        "Zoho CRM",
		Description: "CRM Zoho",
		Command:     "npx",
		Args:        []string{"-y", "zoho-crm-mcp-server"},
		Env:         map[string]string{"ZOHO_CLIENT_ID": "", "ZOHO_CLIENT_SECRET": ""},
		IsEnabled:   false,
	}

	// Monday.com
	cc.mcpServers["monday"] = &MCPServer{
		Name:        "Monday.com",
		Description: "Work OS Monday",
		Command:     "npx",
		Args:        []string{"-y", "monday-mcp-server"},
		Env:         map[string]string{"MONDAY_API_KEY": ""},
		IsEnabled:   false,
	}

	// Airtable
	cc.mcpServers["airtable"] = &MCPServer{
		Name:        "Airtable",
		Description: "Database/Spreadsheet Airtable",
		Command:     "npx",
		Args:        []string{"-y", "airtable-mcp-server"},
		Env:         map[string]string{"AIRTABLE_API_KEY": ""},
		IsEnabled:   false,
	}

	// ClickUp
	cc.mcpServers["clickup"] = &MCPServer{
		Name:        "ClickUp",
		Description: "Gerenciamento de projetos",
		Command:     "npx",
		Args:        []string{"-y", "clickup-mcp-server"},
		Env:         map[string]string{"CLICKUP_API_KEY": ""},
		IsEnabled:   false,
	}

	// Basecamp
	cc.mcpServers["basecamp"] = &MCPServer{
		Name:        "Basecamp",
		Description: "Gerenciamento de projetos",
		Command:     "npx",
		Args:        []string{"-y", "basecamp-mcp-server"},
		Env:         map[string]string{"BASECAMP_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// Zendesk
	cc.mcpServers["zendesk"] = &MCPServer{
		Name:        "Zendesk",
		Description: "Suporte ao cliente",
		Command:     "npx",
		Args:        []string{"-y", "zendesk-mcp-server"},
		Env:         map[string]string{"ZENDESK_SUBDOMAIN": "", "ZENDESK_EMAIL": "", "ZENDESK_TOKEN": ""},
		IsEnabled:   false,
	}

	// Freshdesk
	cc.mcpServers["freshdesk"] = &MCPServer{
		Name:        "Freshdesk",
		Description: "Help desk Freshworks",
		Command:     "npx",
		Args:        []string{"-y", "freshdesk-mcp-server"},
		Env:         map[string]string{"FRESHDESK_DOMAIN": "", "FRESHDESK_API_KEY": ""},
		IsEnabled:   false,
	}

	// Intercom
	cc.mcpServers["intercom"] = &MCPServer{
		Name:        "Intercom",
		Description: "Customer messaging",
		Command:     "npx",
		Args:        []string{"-y", "intercom-mcp-server"},
		Env:         map[string]string{"INTERCOM_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// Drift
	cc.mcpServers["drift"] = &MCPServer{
		Name:        "Drift",
		Description: "Conversational marketing",
		Command:     "npx",
		Args:        []string{"-y", "drift-mcp-server"},
		IsEnabled:   false,
	}

	// Crisp
	cc.mcpServers["crisp"] = &MCPServer{
		Name:        "Crisp",
		Description: "Customer messaging",
		Command:     "npx",
		Args:        []string{"-y", "crisp-mcp-server"},
		Env:         map[string]string{"CRISP_WEBSITE_ID": "", "CRISP_TOKEN_ID": "", "CRISP_TOKEN_KEY": ""},
		IsEnabled:   false,
	}

	// Mailchimp
	cc.mcpServers["mailchimp"] = &MCPServer{
		Name:        "Mailchimp",
		Description: "Email marketing",
		Command:     "npx",
		Args:        []string{"-y", "mailchimp-mcp-server"},
		Env:         map[string]string{"MAILCHIMP_API_KEY": ""},
		IsEnabled:   false,
	}

	// SendGrid
	cc.mcpServers["sendgrid"] = &MCPServer{
		Name:        "SendGrid",
		Description: "Email delivery",
		Command:     "npx",
		Args:        []string{"-y", "sendgrid-mcp-server"},
		Env:         map[string]string{"SENDGRID_API_KEY": ""},
		IsEnabled:   false,
	}

	// Mailgun
	cc.mcpServers["mailgun"] = &MCPServer{
		Name:        "Mailgun",
		Description: "Email API",
		Command:     "npx",
		Args:        []string{"-y", "mailgun-mcp-server"},
		Env:         map[string]string{"MAILGUN_API_KEY": "", "MAILGUN_DOMAIN": ""},
		IsEnabled:   false,
	}

	// Postmark
	cc.mcpServers["postmark"] = &MCPServer{
		Name:        "Postmark",
		Description: "Transactional email",
		Command:     "npx",
		Args:        []string{"-y", "postmark-mcp-server"},
		Env:         map[string]string{"POSTMARK_SERVER_TOKEN": ""},
		IsEnabled:   false,
	}

	// ConvertKit
	cc.mcpServers["convertkit"] = &MCPServer{
		Name:        "ConvertKit",
		Description: "Email para creators",
		Command:     "npx",
		Args:        []string{"-y", "convertkit-mcp-server"},
		Env:         map[string]string{"CONVERTKIT_API_SECRET": ""},
		IsEnabled:   false,
	}

	// ActiveCampaign
	cc.mcpServers["activecampaign"] = &MCPServer{
		Name:        "ActiveCampaign",
		Description: "Marketing automation",
		Command:     "npx",
		Args:        []string{"-y", "activecampaign-mcp-server"},
		Env:         map[string]string{"AC_API_URL": "", "AC_API_KEY": ""},
		IsEnabled:   false,
	}

	// Drip
	cc.mcpServers["drip"] = &MCPServer{
		Name:        "Drip",
		Description: "Ecommerce CRM",
		Command:     "npx",
		Args:        []string{"-y", "drip-mcp-server"},
		Env:         map[string]string{"DRIP_API_KEY": "", "DRIP_ACCOUNT_ID": ""},
		IsEnabled:   false,
	}

	// Calendly
	cc.mcpServers["calendly"] = &MCPServer{
		Name:        "Calendly",
		Description: "Agendamento de reuniões",
		Command:     "npx",
		Args:        []string{"-y", "calendly-mcp-server"},
		Env:         map[string]string{"CALENDLY_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// Cal.com
	cc.mcpServers["cal-com"] = &MCPServer{
		Name:        "Cal.com",
		Description: "Scheduling open-source",
		Command:     "npx",
		Args:        []string{"-y", "cal-com-mcp-server"},
		Env:         map[string]string{"CAL_API_KEY": ""},
		IsEnabled:   false,
	}

	// Doodle
	cc.mcpServers["doodle"] = &MCPServer{
		Name:        "Doodle",
		Description: "Meeting scheduling",
		Command:     "npx",
		Args:        []string{"-y", "doodle-mcp-server"},
		IsEnabled:   false,
	}

	// Zoom
	cc.mcpServers["zoom"] = &MCPServer{
		Name:        "Zoom",
		Description: "Video conferencing",
		Command:     "npx",
		Args:        []string{"-y", "zoom-mcp-server"},
		Env:         map[string]string{"ZOOM_CLIENT_ID": "", "ZOOM_CLIENT_SECRET": ""},
		IsEnabled:   false,
	}

	// Google Meet
	cc.mcpServers["google-meet"] = &MCPServer{
		Name:        "Google Meet",
		Description: "Video calls Google",
		Command:     "npx",
		Args:        []string{"-y", "google-meet-mcp-server"},
		IsEnabled:   false,
	}

	// Microsoft Teams
	cc.mcpServers["ms-teams"] = &MCPServer{
		Name:        "Microsoft Teams",
		Description: "Colaboração Teams",
		Command:     "npx",
		Args:        []string{"-y", "ms-teams-mcp-server"},
		Env:         map[string]string{"TEAMS_CLIENT_ID": "", "TEAMS_CLIENT_SECRET": ""},
		IsEnabled:   false,
	}

	// Webex
	cc.mcpServers["webex"] = &MCPServer{
		Name:        "Cisco Webex",
		Description: "Video conferencing",
		Command:     "npx",
		Args:        []string{"-y", "webex-mcp-server"},
		Env:         map[string]string{"WEBEX_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// GoTo Meeting
	cc.mcpServers["goto-meeting"] = &MCPServer{
		Name:        "GoTo Meeting",
		Description: "Video meetings",
		Command:     "npx",
		Args:        []string{"-y", "goto-meeting-mcp-server"},
		IsEnabled:   false,
	}

	// Loom
	cc.mcpServers["loom"] = &MCPServer{
		Name:        "Loom",
		Description: "Video messaging",
		Command:     "npx",
		Args:        []string{"-y", "loom-mcp-server"},
		Env:         map[string]string{"LOOM_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// Typeform
	cc.mcpServers["typeform"] = &MCPServer{
		Name:        "Typeform",
		Description: "Formulários interativos",
		Command:     "npx",
		Args:        []string{"-y", "typeform-mcp-server"},
		Env:         map[string]string{"TYPEFORM_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// Google Forms
	cc.mcpServers["google-forms"] = &MCPServer{
		Name:        "Google Forms",
		Description: "Formulários Google",
		Command:     "npx",
		Args:        []string{"-y", "google-forms-mcp-server"},
		IsEnabled:   false,
	}

	// SurveyMonkey
	cc.mcpServers["surveymonkey"] = &MCPServer{
		Name:        "SurveyMonkey",
		Description: "Pesquisas online",
		Command:     "npx",
		Args:        []string{"-y", "surveymonkey-mcp-server"},
		Env:         map[string]string{"SURVEYMONKEY_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// JotForm
	cc.mcpServers["jotform"] = &MCPServer{
		Name:        "JotForm",
		Description: "Criador de formulários",
		Command:     "npx",
		Args:        []string{"-y", "jotform-mcp-server"},
		Env:         map[string]string{"JOTFORM_API_KEY": ""},
		IsEnabled:   false,
	}

	// DocSend
	cc.mcpServers["docsend"] = &MCPServer{
		Name:        "DocSend",
		Description: "Compartilhamento de docs",
		Command:     "npx",
		Args:        []string{"-y", "docsend-mcp-server"},
		IsEnabled:   false,
	}

	// Box
	cc.mcpServers["box"] = &MCPServer{
		Name:        "Box",
		Description: "Cloud content management",
		Command:     "npx",
		Args:        []string{"-y", "box-mcp-server"},
		Env:         map[string]string{"BOX_CLIENT_ID": "", "BOX_CLIENT_SECRET": ""},
		IsEnabled:   false,
	}

	// Dropbox
	cc.mcpServers["dropbox"] = &MCPServer{
		Name:        "Dropbox",
		Description: "Cloud storage",
		Command:     "npx",
		Args:        []string{"-y", "dropbox-mcp-server"},
		Env:         map[string]string{"DROPBOX_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// RD Station (Brasil)
	cc.mcpServers["rd-station"] = &MCPServer{
		Name:        "RD Station",
		Description: "Marketing digital Brasil",
		Command:     "npx",
		Args:        []string{"-y", "rd-station-mcp-server"},
		Env:         map[string]string{"RD_CLIENT_ID": "", "RD_CLIENT_SECRET": ""},
		IsEnabled:   false,
	}
}

// RegisterHRMCP registra servidores de RH
func (cc *ClaudeCode) RegisterHRMCP() {
	// Workday
	cc.mcpServers["workday"] = &MCPServer{
		Name:        "Workday",
		Description: "HCM Workday",
		Command:     "npx",
		Args:        []string{"-y", "workday-mcp-server"},
		IsEnabled:   false,
	}

	// BambooHR
	cc.mcpServers["bamboohr"] = &MCPServer{
		Name:        "BambooHR",
		Description: "HR Software",
		Command:     "npx",
		Args:        []string{"-y", "bamboohr-mcp-server"},
		Env:         map[string]string{"BAMBOOHR_SUBDOMAIN": "", "BAMBOOHR_API_KEY": ""},
		IsEnabled:   false,
	}

	// Gusto
	cc.mcpServers["gusto"] = &MCPServer{
		Name:        "Gusto",
		Description: "Payroll e benefits",
		Command:     "npx",
		Args:        []string{"-y", "gusto-mcp-server"},
		Env:         map[string]string{"GUSTO_CLIENT_ID": "", "GUSTO_CLIENT_SECRET": ""},
		IsEnabled:   false,
	}

	// Rippling
	cc.mcpServers["rippling"] = &MCPServer{
		Name:        "Rippling",
		Description: "HR, IT, Finance",
		Command:     "npx",
		Args:        []string{"-y", "rippling-mcp-server"},
		IsEnabled:   false,
	}

	// Lever
	cc.mcpServers["lever"] = &MCPServer{
		Name:        "Lever",
		Description: "ATS recruiting",
		Command:     "npx",
		Args:        []string{"-y", "lever-mcp-server"},
		Env:         map[string]string{"LEVER_API_KEY": ""},
		IsEnabled:   false,
	}

	// Greenhouse
	cc.mcpServers["greenhouse"] = &MCPServer{
		Name:        "Greenhouse",
		Description: "Recruiting software",
		Command:     "npx",
		Args:        []string{"-y", "greenhouse-mcp-server"},
		Env:         map[string]string{"GREENHOUSE_API_KEY": ""},
		IsEnabled:   false,
	}

	// Deel
	cc.mcpServers["deel"] = &MCPServer{
		Name:        "Deel",
		Description: "Global payroll",
		Command:     "npx",
		Args:        []string{"-y", "deel-mcp-server"},
		Env:         map[string]string{"DEEL_API_KEY": ""},
		IsEnabled:   false,
	}

	// Remote.com
	cc.mcpServers["remote"] = &MCPServer{
		Name:        "Remote.com",
		Description: "Global HR platform",
		Command:     "npx",
		Args:        []string{"-y", "remote-mcp-server"},
		IsEnabled:   false,
	}

	// Lattice
	cc.mcpServers["lattice"] = &MCPServer{
		Name:        "Lattice",
		Description: "Performance management",
		Command:     "npx",
		Args:        []string{"-y", "lattice-mcp-server"},
		IsEnabled:   false,
	}

	// 15Five
	cc.mcpServers["15five"] = &MCPServer{
		Name:        "15Five",
		Description: "Performance e engagement",
		Command:     "npx",
		Args:        []string{"-y", "15five-mcp-server"},
		IsEnabled:   false,
	}

	// Culture Amp
	cc.mcpServers["culture-amp"] = &MCPServer{
		Name:        "Culture Amp",
		Description: "Employee experience",
		Command:     "npx",
		Args:        []string{"-y", "culture-amp-mcp-server"},
		IsEnabled:   false,
	}

	// Gupy (Brasil)
	cc.mcpServers["gupy"] = &MCPServer{
		Name:        "Gupy",
		Description: "Recrutamento Brasil",
		Command:     "npx",
		Args:        []string{"-y", "gupy-mcp-server"},
		IsEnabled:   false,
	}
}

// RegisterDevOpsMCP registra servidores DevOps adicionais
func (cc *ClaudeCode) RegisterDevOpsMCP() {
	// Já tem: Docker, Kubernetes, AWS, Azure, GCP

	// Terraform
	cc.mcpServers["terraform"] = &MCPServer{
		Name:        "Terraform",
		Description: "Infrastructure as Code",
		Command:     "npx",
		Args:        []string{"-y", "terraform-mcp-server"},
		IsEnabled:   false,
	}

	// Pulumi
	cc.mcpServers["pulumi"] = &MCPServer{
		Name:        "Pulumi",
		Description: "Modern IaC",
		Command:     "npx",
		Args:        []string{"-y", "pulumi-mcp-server"},
		Env:         map[string]string{"PULUMI_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// Ansible
	cc.mcpServers["ansible"] = &MCPServer{
		Name:        "Ansible",
		Description: "Automation platform",
		Command:     "npx",
		Args:        []string{"-y", "ansible-mcp-server"},
		IsEnabled:   false,
	}

	// Jenkins
	cc.mcpServers["jenkins"] = &MCPServer{
		Name:        "Jenkins",
		Description: "CI/CD server",
		Command:     "npx",
		Args:        []string{"-y", "jenkins-mcp-server"},
		Env:         map[string]string{"JENKINS_URL": "", "JENKINS_USER": "", "JENKINS_TOKEN": ""},
		IsEnabled:   false,
	}

	// CircleCI
	cc.mcpServers["circleci"] = &MCPServer{
		Name:        "CircleCI",
		Description: "CI/CD platform",
		Command:     "npx",
		Args:        []string{"-y", "circleci-mcp-server"},
		Env:         map[string]string{"CIRCLECI_TOKEN": ""},
		IsEnabled:   false,
	}

	// Travis CI
	cc.mcpServers["travis"] = &MCPServer{
		Name:        "Travis CI",
		Description: "Continuous integration",
		Command:     "npx",
		Args:        []string{"-y", "travis-mcp-server"},
		Env:         map[string]string{"TRAVIS_TOKEN": ""},
		IsEnabled:   false,
	}

	// GitHub Actions
	cc.mcpServers["github-actions"] = &MCPServer{
		Name:        "GitHub Actions",
		Description: "CI/CD GitHub",
		Command:     "npx",
		Args:        []string{"-y", "github-actions-mcp-server"},
		Env:         map[string]string{"GITHUB_TOKEN": ""},
		IsEnabled:   false,
	}

	// GitLab CI
	cc.mcpServers["gitlab-ci"] = &MCPServer{
		Name:        "GitLab CI/CD",
		Description: "CI/CD GitLab",
		Command:     "npx",
		Args:        []string{"-y", "gitlab-ci-mcp-server"},
		Env:         map[string]string{"GITLAB_TOKEN": ""},
		IsEnabled:   false,
	}

	// Bitbucket Pipelines
	cc.mcpServers["bitbucket-pipelines"] = &MCPServer{
		Name:        "Bitbucket Pipelines",
		Description: "CI/CD Atlassian",
		Command:     "npx",
		Args:        []string{"-y", "bitbucket-pipelines-mcp-server"},
		IsEnabled:   false,
	}

	// ArgoCD
	cc.mcpServers["argocd"] = &MCPServer{
		Name:        "ArgoCD",
		Description: "GitOps for Kubernetes",
		Command:     "npx",
		Args:        []string{"-y", "argocd-mcp-server"},
		Env:         map[string]string{"ARGOCD_SERVER": "", "ARGOCD_TOKEN": ""},
		IsEnabled:   false,
	}

	// Flux
	cc.mcpServers["flux"] = &MCPServer{
		Name:        "Flux",
		Description: "GitOps toolkit",
		Command:     "npx",
		Args:        []string{"-y", "flux-mcp-server"},
		IsEnabled:   false,
	}

	// Helm
	cc.mcpServers["helm"] = &MCPServer{
		Name:        "Helm",
		Description: "Kubernetes package manager",
		Command:     "npx",
		Args:        []string{"-y", "helm-mcp-server"},
		IsEnabled:   false,
	}

	// Prometheus
	cc.mcpServers["prometheus"] = &MCPServer{
		Name:        "Prometheus",
		Description: "Monitoring system",
		Command:     "npx",
		Args:        []string{"-y", "prometheus-mcp-server"},
		Env:         map[string]string{"PROMETHEUS_URL": ""},
		IsEnabled:   false,
	}

	// Grafana
	cc.mcpServers["grafana"] = &MCPServer{
		Name:        "Grafana",
		Description: "Observability platform",
		Command:     "npx",
		Args:        []string{"-y", "grafana-mcp-server"},
		Env:         map[string]string{"GRAFANA_URL": "", "GRAFANA_API_KEY": ""},
		IsEnabled:   false,
	}

	// Datadog
	cc.mcpServers["datadog"] = &MCPServer{
		Name:        "Datadog",
		Description: "Cloud monitoring",
		Command:     "npx",
		Args:        []string{"-y", "datadog-mcp-server"},
		Env:         map[string]string{"DD_API_KEY": "", "DD_APP_KEY": ""},
		IsEnabled:   false,
	}

	// New Relic
	cc.mcpServers["newrelic"] = &MCPServer{
		Name:        "New Relic",
		Description: "Observability platform",
		Command:     "npx",
		Args:        []string{"-y", "newrelic-mcp-server"},
		Env:         map[string]string{"NEW_RELIC_API_KEY": ""},
		IsEnabled:   false,
	}

	// Splunk
	cc.mcpServers["splunk"] = &MCPServer{
		Name:        "Splunk",
		Description: "Log analysis",
		Command:     "npx",
		Args:        []string{"-y", "splunk-mcp-server"},
		Env:         map[string]string{"SPLUNK_HOST": "", "SPLUNK_TOKEN": ""},
		IsEnabled:   false,
	}

	// Elastic/ELK
	cc.mcpServers["elasticsearch"] = &MCPServer{
		Name:        "Elasticsearch",
		Description: "Search and analytics",
		Command:     "npx",
		Args:        []string{"-y", "elasticsearch-mcp-server"},
		Env:         map[string]string{"ELASTICSEARCH_URL": "", "ELASTICSEARCH_API_KEY": ""},
		IsEnabled:   false,
	}

	// PagerDuty
	cc.mcpServers["pagerduty"] = &MCPServer{
		Name:        "PagerDuty",
		Description: "Incident management",
		Command:     "npx",
		Args:        []string{"-y", "pagerduty-mcp-server"},
		Env:         map[string]string{"PAGERDUTY_API_KEY": ""},
		IsEnabled:   false,
	}

	// OpsGenie
	cc.mcpServers["opsgenie"] = &MCPServer{
		Name:        "OpsGenie",
		Description: "Alerting and on-call",
		Command:     "npx",
		Args:        []string{"-y", "opsgenie-mcp-server"},
		Env:         map[string]string{"OPSGENIE_API_KEY": ""},
		IsEnabled:   false,
	}

	// StatusPage
	cc.mcpServers["statuspage"] = &MCPServer{
		Name:        "Statuspage",
		Description: "Status page hosting",
		Command:     "npx",
		Args:        []string{"-y", "statuspage-mcp-server"},
		Env:         map[string]string{"STATUSPAGE_API_KEY": ""},
		IsEnabled:   false,
	}

	// Sentry
	cc.mcpServers["sentry"] = &MCPServer{
		Name:        "Sentry",
		Description: "Error tracking",
		Command:     "npx",
		Args:        []string{"-y", "sentry-mcp-server"},
		Env:         map[string]string{"SENTRY_AUTH_TOKEN": "", "SENTRY_ORG": ""},
		IsEnabled:   false,
	}

	// Bugsnag
	cc.mcpServers["bugsnag"] = &MCPServer{
		Name:        "Bugsnag",
		Description: "Error monitoring",
		Command:     "npx",
		Args:        []string{"-y", "bugsnag-mcp-server"},
		Env:         map[string]string{"BUGSNAG_API_KEY": ""},
		IsEnabled:   false,
	}

	// Rollbar
	cc.mcpServers["rollbar"] = &MCPServer{
		Name:        "Rollbar",
		Description: "Error tracking",
		Command:     "npx",
		Args:        []string{"-y", "rollbar-mcp-server"},
		Env:         map[string]string{"ROLLBAR_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// LaunchDarkly
	cc.mcpServers["launchdarkly"] = &MCPServer{
		Name:        "LaunchDarkly",
		Description: "Feature flags",
		Command:     "npx",
		Args:        []string{"-y", "launchdarkly-mcp-server"},
		Env:         map[string]string{"LAUNCHDARKLY_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// Split.io
	cc.mcpServers["split"] = &MCPServer{
		Name:        "Split.io",
		Description: "Feature delivery",
		Command:     "npx",
		Args:        []string{"-y", "split-mcp-server"},
		Env:         map[string]string{"SPLIT_API_KEY": ""},
		IsEnabled:   false,
	}

	// Vault
	cc.mcpServers["vault"] = &MCPServer{
		Name:        "HashiCorp Vault",
		Description: "Secrets management",
		Command:     "npx",
		Args:        []string{"-y", "vault-mcp-server"},
		Env:         map[string]string{"VAULT_ADDR": "", "VAULT_TOKEN": ""},
		IsEnabled:   false,
	}

	// Snyk
	cc.mcpServers["snyk"] = &MCPServer{
		Name:        "Snyk",
		Description: "Security scanning",
		Command:     "npx",
		Args:        []string{"-y", "snyk-mcp-server"},
		Env:         map[string]string{"SNYK_TOKEN": ""},
		IsEnabled:   false,
	}

	// SonarQube
	cc.mcpServers["sonarqube"] = &MCPServer{
		Name:        "SonarQube",
		Description: "Code quality",
		Command:     "npx",
		Args:        []string{"-y", "sonarqube-mcp-server"},
		Env:         map[string]string{"SONAR_HOST_URL": "", "SONAR_TOKEN": ""},
		IsEnabled:   false,
	}

	// Codecov
	cc.mcpServers["codecov"] = &MCPServer{
		Name:        "Codecov",
		Description: "Code coverage",
		Command:     "npx",
		Args:        []string{"-y", "codecov-mcp-server"},
		Env:         map[string]string{"CODECOV_TOKEN": ""},
		IsEnabled:   false,
	}
}

// RegisterAllMCP registra todos os servidores MCP
func (cc *ClaudeCode) RegisterAllMCP() {
	cc.RegisterSmartHomeMCP()
	cc.RegisterHealthMCP()
	cc.RegisterTravelMCP()
	cc.RegisterShoppingMCP()
	cc.RegisterSocialMediaMCP()
	cc.RegisterGamingMCP()
	cc.RegisterEducationMCP()
	cc.RegisterNewsMCP()
	cc.RegisterWeatherMCP()
	cc.RegisterSportsMCP()
	cc.RegisterFinanceMCP()
	cc.RegisterLegalMCP()
	cc.RegisterRealEstateMCP()
	cc.RegisterAutomotiveMCP()
	cc.RegisterEntertainmentMCP()
	cc.RegisterDesignMCP()
	cc.RegisterBusinessMCP()
	cc.RegisterHRMCP()
	cc.RegisterDevOpsMCP()
}
