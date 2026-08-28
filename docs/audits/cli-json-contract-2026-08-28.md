# CLI JSON contract repair audit — 2026-08-28

Repository revision: `123b252bdc96dad8ca7397fd205127a4fd554456` plus the audited repair worktree
Contract version: `1.0.0`
Contract released: `false`
Audit standard: `docs/cli-contract-audit.md` version `1.0.0`
Executable commands: `108`
Input schemas: `108`
Response schemas: `108`
Shared schemas: `10`
Pre-existing dirty files: none
Verdict: **AUTHORITATIVE: PASS**
Findings: `0`

This repair audit reuses the frozen scope and standard of the 2026-08-27 audit. The worktree was clean on `main` before that audit; the repaired contract lock SHA-256 is `124f2aadfac96ae4aebfb084e8418d2d935f436c7284f77394e0e7919fdad212`.

## Inventory

Exact set comparison: `E = C = I = R = D = B = M`, with 108 members in every set and zero symmetric-difference members. `M` was checked against the command tree at `docs/CLI.md:5` and the exact or explicit shared command sections beginning at `docs/CLI.md:842`. Grouped entries are the condition mutations, group mutations, auth/project quota-project trios, and four completion operations.

Capability path, schema-reference, and DTO-registration equality is enforced by `TestEveryExecutableCommandHasCapabilityAndPublishedSchemas`; behavior-manifest equality is frozen by `contract_v1_capabilities_detailed.golden.json`; `TestEveryExecutableCommandHasDocumentationInventoryEntry` compares `M` with `E`; and `contract_v1_audit_evidence.golden.json` covers the complete command-by-test-class matrix.

### Shared schemas

| File | `$id` |
| --- | --- |
| `schemas/cli/1.0.0/capability.schema.json` | `urn:fbrcm:schema:cli:1.0.0:capability` |
| `schemas/cli/1.0.0/envelope.schema.json` | `urn:fbrcm:schema:cli:1.0.0:envelope` |
| `schemas/cli/1.0.0/error.schema.json` | `urn:fbrcm:schema:cli:1.0.0:error` |
| `schemas/cli/1.0.0/semantic.schema.json` | `urn:fbrcm:schema:cli:1.0.0:semantic` |
| `schemas/cli/1.0.0/stdin.credentials.schema.json` | `urn:fbrcm:schema:cli:1.0.0:stdin:credentials` |
| `schemas/cli/1.0.0/stdin.oauth_credentials.schema.json` | `urn:fbrcm:schema:cli:1.0.0:stdin:oauth_credentials` |
| `schemas/cli/1.0.0/stdin.remote_config.schema.json` | `urn:fbrcm:schema:cli:1.0.0:stdin:remote_config` |
| `schemas/cli/1.0.0/stdin.remote_config_import.schema.json` | `urn:fbrcm:schema:cli:1.0.0:stdin:remote_config_import` |
| `schemas/cli/1.0.0/stdin.service_account_credentials.schema.json` | `urn:fbrcm:schema:cli:1.0.0:stdin:service_account_credentials` |
| `schemas/cli/1.0.0/stdin.theme.schema.json` | `urn:fbrcm:schema:cli:1.0.0:stdin:theme` |

### Command schemas

Every command schema file and `$id` is recorded below; this is also the `I` and `R` inventory.

| Command | Input file / `$id` | Response file / `$id` |
| --- | --- | --- |
| `add` | `add.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:add:input` | `add.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:add:response` |
| `auth.add.gcloud` | `auth_add_gcloud.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:auth.add.gcloud:input` | `auth_add_gcloud.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:auth.add.gcloud:response` |
| `auth.add.oauth` | `auth_add_oauth.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:auth.add.oauth:input` | `auth_add_oauth.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:auth.add.oauth:response` |
| `auth.add.service-account` | `auth_add_service-account.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:auth.add.service-account:input` | `auth_add_service-account.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:auth.add.service-account:response` |
| `auth.bind` | `auth_bind.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:auth.bind:input` | `auth_bind.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:auth.bind:response` |
| `auth.delete` | `auth_delete.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:auth.delete:input` | `auth_delete.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:auth.delete:response` |
| `auth.list` | `auth_list.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:auth.list:input` | `auth_list.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:auth.list:response` |
| `auth.login` | `auth_login.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:auth.login:input` | `auth_login.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:auth.login:response` |
| `auth.path` | `auth_path.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:auth.path:input` | `auth_path.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:auth.path:response` |
| `auth.quota-project.set` | `auth_quota-project_set.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.set:input` | `auth_quota-project_set.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.set:response` |
| `auth.quota-project.show` | `auth_quota-project_show.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.show:input` | `auth_quota-project_show.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.show:response` |
| `auth.quota-project.unset` | `auth_quota-project_unset.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.unset:input` | `auth_quota-project_unset.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.unset:response` |
| `cache.clear` | `cache_clear.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:cache.clear:input` | `cache_clear.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:cache.clear:response` |
| `cache.list` | `cache_list.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:cache.list:input` | `cache_list.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:cache.list:response` |
| `cache.path` | `cache_path.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:cache.path:input` | `cache_path.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:cache.path:response` |
| `capabilities` | `capabilities.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:capabilities:input` | `capabilities.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:capabilities:response` |
| `completion.bash` | `completion_bash.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:completion.bash:input` | `completion_bash.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:completion.bash:response` |
| `completion.fish` | `completion_fish.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:completion.fish:input` | `completion_fish.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:completion.fish:response` |
| `completion.powershell` | `completion_powershell.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:completion.powershell:input` | `completion_powershell.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:completion.powershell:response` |
| `completion.zsh` | `completion_zsh.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:completion.zsh:input` | `completion_zsh.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:completion.zsh:response` |
| `conditions.add` | `conditions_add.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:conditions.add:input` | `conditions_add.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:conditions.add:response` |
| `conditions.delete` | `conditions_delete.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:conditions.delete:input` | `conditions_delete.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:conditions.delete:response` |
| `conditions.edit` | `conditions_edit.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:conditions.edit:input` | `conditions_edit.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:conditions.edit:response` |
| `conditions.list` | `conditions_list.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:conditions.list:input` | `conditions_list.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:conditions.list:response` |
| `conditions.move` | `conditions_move.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:conditions.move:input` | `conditions_move.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:conditions.move:response` |
| `conditions.rename` | `conditions_rename.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:conditions.rename:input` | `conditions_rename.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:conditions.rename:response` |
| `conditions.show` | `conditions_show.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:conditions.show:input` | `conditions_show.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:conditions.show:response` |
| `conditions.validate` | `conditions_validate.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:conditions.validate:input` | `conditions_validate.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:conditions.validate:response` |
| `config.edit` | `config_edit.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:config.edit:input` | `config_edit.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:config.edit:response` |
| `config.path` | `config_path.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:config.path:input` | `config_path.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:config.path:response` |
| `config.reset` | `config_reset.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:config.reset:input` | `config_reset.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:config.reset:response` |
| `config.set` | `config_set.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:config.set:input` | `config_set.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:config.set:response` |
| `config.show` | `config_show.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:config.show:input` | `config_show.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:config.show:response` |
| `config.validate` | `config_validate.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:config.validate:input` | `config_validate.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:config.validate:response` |
| `delete` | `delete.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:delete:input` | `delete.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:delete:response` |
| `doctor` | `doctor.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:doctor:input` | `doctor.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:doctor:response` |
| `draft.change-note` | `draft_change-note.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:draft.change-note:input` | `draft_change-note.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:draft.change-note:response` |
| `draft.diff` | `draft_diff.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:draft.diff:input` | `draft_diff.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:draft.diff:response` |
| `draft.discard` | `draft_discard.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:draft.discard:input` | `draft_discard.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:draft.discard:response` |
| `draft.list` | `draft_list.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:draft.list:input` | `draft_list.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:draft.list:response` |
| `draft.path` | `draft_path.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:draft.path:input` | `draft_path.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:draft.path:response` |
| `draft.publish` | `draft_publish.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:draft.publish:input` | `draft_publish.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:draft.publish:response` |
| `draft.show` | `draft_show.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:draft.show:input` | `draft_show.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:draft.show:response` |
| `duplicate` | `duplicate.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:duplicate:input` | `duplicate.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:duplicate:response` |
| `experiments.delete` | `experiments_delete.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:experiments.delete:input` | `experiments_delete.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:experiments.delete:response` |
| `experiments.list` | `experiments_list.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:experiments.list:input` | `experiments_list.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:experiments.list:response` |
| `experiments.show` | `experiments_show.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:experiments.show:input` | `experiments_show.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:experiments.show:response` |
| `get` | `get.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:get:input` | `get.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:get:response` |
| `groups.add` | `groups_add.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:groups.add:input` | `groups_add.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:groups.add:response` |
| `groups.delete` | `groups_delete.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:groups.delete:input` | `groups_delete.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:groups.delete:response` |
| `groups.edit` | `groups_edit.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:groups.edit:input` | `groups_edit.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:groups.edit:response` |
| `groups.list` | `groups_list.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:groups.list:input` | `groups_list.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:groups.list:response` |
| `groups.rename` | `groups_rename.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:groups.rename:input` | `groups_rename.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:groups.rename:response` |
| `help` | `help.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:help:input` | `help.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:help:response` |
| `hooks.fingerprint` | `hooks_fingerprint.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:hooks.fingerprint:input` | `hooks_fingerprint.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:hooks.fingerprint:response` |
| `hooks.status` | `hooks_status.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:hooks.status:input` | `hooks_status.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:hooks.status:response` |
| `hooks.trust` | `hooks_trust.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:hooks.trust:input` | `hooks_trust.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:hooks.trust:response` |
| `hooks.untrust` | `hooks_untrust.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:hooks.untrust:input` | `hooks_untrust.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:hooks.untrust:response` |
| `personalizations.list` | `personalizations_list.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:personalizations.list:input` | `personalizations_list.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:personalizations.list:response` |
| `personalizations.show` | `personalizations_show.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:personalizations.show:input` | `personalizations_show.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:personalizations.show:response` |
| `profile` | `profile.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:profile:input` | `profile.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:profile:response` |
| `profile.delete` | `profile_delete.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:profile.delete:input` | `profile_delete.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:profile.delete:response` |
| `profile.list` | `profile_list.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:profile.list:input` | `profile_list.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:profile.list:response` |
| `profile.path` | `profile_path.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:profile.path:input` | `profile_path.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:profile.path:response` |
| `profile.rename` | `profile_rename.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:profile.rename:input` | `profile_rename.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:profile.rename:response` |
| `profile.switch` | `profile_switch.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:profile.switch:input` | `profile_switch.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:profile.switch:response` |
| `project.defaults` | `project_defaults.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:project.defaults:input` | `project_defaults.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:project.defaults:response` |
| `project.export` | `project_export.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:project.export:input` | `project_export.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:project.export:response` |
| `project.import` | `project_import.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:project.import:input` | `project_import.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:project.import:response` |
| `project.open` | `project_open.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:project.open:input` | `project_open.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:project.open:response` |
| `project.quota-project.set` | `project_quota-project_set.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.set:input` | `project_quota-project_set.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.set:response` |
| `project.quota-project.show` | `project_quota-project_show.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.show:input` | `project_quota-project_show.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.show:response` |
| `project.quota-project.unset` | `project_quota-project_unset.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.unset:input` | `project_quota-project_unset.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.unset:response` |
| `project.show` | `project_show.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:project.show:input` | `project_show.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:project.show:response` |
| `project.templates.set` | `project_templates_set.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:project.templates.set:input` | `project_templates_set.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:project.templates.set:response` |
| `project.templates.show` | `project_templates_show.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:project.templates.show:input` | `project_templates_show.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:project.templates.show:response` |
| `projects.aliases.import` | `projects_aliases_import.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.import:input` | `projects_aliases_import.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.import:response` |
| `projects.aliases.list` | `projects_aliases_list.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.list:input` | `projects_aliases_list.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.list:response` |
| `projects.aliases.remove` | `projects_aliases_remove.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.remove:input` | `projects_aliases_remove.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.remove:response` |
| `projects.aliases.set` | `projects_aliases_set.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.set:input` | `projects_aliases_set.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.set:response` |
| `projects.diff` | `projects_diff.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:projects.diff:input` | `projects_diff.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:projects.diff:response` |
| `projects.forget` | `projects_forget.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:projects.forget:input` | `projects_forget.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:projects.forget:response` |
| `projects.list` | `projects_list.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:projects.list:input` | `projects_list.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:projects.list:response` |
| `projects.path` | `projects_path.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:projects.path:input` | `projects_path.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:projects.path:response` |
| `projects.promote` | `projects_promote.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:projects.promote:input` | `projects_promote.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:projects.promote:response` |
| `projects.reset` | `projects_reset.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:projects.reset:input` | `projects_reset.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:projects.reset:response` |
| `projects.update` | `projects_update.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:projects.update:input` | `projects_update.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:projects.update:response` |
| `rollouts.delete` | `rollouts_delete.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:rollouts.delete:input` | `rollouts_delete.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:rollouts.delete:response` |
| `rollouts.list` | `rollouts_list.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:rollouts.list:input` | `rollouts_list.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:rollouts.list:response` |
| `rollouts.show` | `rollouts_show.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:rollouts.show:input` | `rollouts_show.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:rollouts.show:response` |
| `root` | `root.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:root:input` | `root.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:root:response` |
| `schema.list` | `schema_list.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:schema.list:input` | `schema_list.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:schema.list:response` |
| `schema.show` | `schema_show.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:schema.show:input` | `schema_show.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:schema.show:response` |
| `theme` | `theme.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:theme:input` | `theme.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:theme:response` |
| `theme.delete` | `theme_delete.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:theme.delete:input` | `theme_delete.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:theme.delete:response` |
| `theme.import` | `theme_import.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:theme.import:input` | `theme_import.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:theme.import:response` |
| `theme.list` | `theme_list.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:theme.list:input` | `theme_list.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:theme.list:response` |
| `theme.path` | `theme_path.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:theme.path:input` | `theme_path.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:theme.path:response` |
| `theme.rename` | `theme_rename.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:theme.rename:input` | `theme_rename.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:theme.rename:response` |
| `theme.reset` | `theme_reset.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:theme.reset:input` | `theme_reset.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:theme.reset:response` |
| `theme.switch` | `theme_switch.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:theme.switch:input` | `theme_switch.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:theme.switch:response` |
| `update` | `update.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:update:input` | `update.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:update:response` |
| `versions.diff` | `versions_diff.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:versions.diff:input` | `versions_diff.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:versions.diff:response` |
| `versions.export` | `versions_export.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:versions.export:input` | `versions_export.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:versions.export:response` |
| `versions.list` | `versions_list.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:versions.list:input` | `versions_list.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:versions.list:response` |
| `versions.restore` | `versions_restore.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:versions.restore:input` | `versions_restore.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:versions.restore:response` |
| `versions.rollback` | `versions_rollback.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:versions.rollback:input` | `versions_rollback.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:versions.rollback:response` |
| `versions.show` | `versions_show.input.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:versions.show:input` | `versions_show.response.schema.json` / `urn:fbrcm:schema:cli:1.0.0:command:versions.show:response` |

All 226 files declare Draft 2020-12, all `$id` values are unique, and compiling every `$id` resolved every reference. No duplicate or dangling IDs/references were found.

## Criterion results

| Result | Criteria |
| --- | --- |
| PASS | `INV-01`–`INV-04`; `ARG-01`–`ARG-07`; `SEL-01`–`SEL-05`; `STDIN-01`–`STDIN-04`; `OUT-01`–`OUT-06`; `ERR-01`–`ERR-06`; `BEH-01`–`BEH-06`; `INT-01`–`INT-04`; `DOC-01`–`DOC-04`; `GEN-01`–`GEN-05` |
| FAIL | None. |

All 51 criteria pass. The repair retested every changed or dependent record and the generated evidence matrix supplies a test ID or justified `N/A` for every section-6 class and command.

## Command audit records

All JSON pointers below are relative to the named generated schema. The generated input schema is the complete machine-readable source for grammar, normalization, selector composition, dependencies, exclusions, and runtime-validation annotations; the detailed capability golden is the complete source for conditional effects and interaction predicates. Repeating those large JSON objects verbatim in this report would create a second contract copy, so every record names the exact authoritative object and summarizes its audited contents.

### `add`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `add`; argv `fbrcm add`. |
| Arguments | `parameter` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/add.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--change-note` (string, default=""); `--description` (string, default=""); `--draft` (bool, default=false, effective_when=1 clause(s)); `--dry-run` (bool, default=false); `--expr` (string, default=""); `--group` (string, default=""); `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--project` (stringArray, default=[], aliases=-p, repeatable); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--type` (string, default=""); `--use-in-app-default` (bool, default=false); `--value` (string, default=""); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/add.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 3 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/add.input.schema.json`: `#/$defs/stateless_get_project_selector/oneOf/0/x-fbrcm-matching`, `#/$defs/stateless_get_project_selector/oneOf/1/x-fbrcm-matching`, `#/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | Modes `json_document`; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:remote_config`; normalized property `schemas/cli/1.0.0/add.input.schema.json#/properties/stdin`; option restrictions are schema-enforced. |
| Success | Registered DTO type(s): `[]shared/rc.RemoteMutationJSONResult | contract.ArtifactData`; structural schema `schemas/cli/1.0.0/add.response.schema.json#/$defs/success_data`; reachable outcome set `success|partial_success|failure`; success variants closed variants `oneOf/0/properties/items/items/allOf/9/then/properties/error/properties/stage=publication|pre_publish_hook; oneOf/0/properties/items/items/allOf/11/then/properties/error/properties/stage=preparation|validation|publication; oneOf/1/$defs/artifact_json_content/properties/parameterGroups/additionalProperties/properties/parameters/additionalProperties/properties/conditionalValues/additionalProperties/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; oneOf/1/$defs/artifact_json_content/properties/parameterGroups/additionalProperties/properties/parameters/additionalProperties/properties/defaultValue/oneOf/0/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; oneOf/1/$defs/artifact_json_content/properties/parameterGroups/additionalProperties/properties/parameters/additionalProperties/properties/valueType=|PARAMETER_VALUE_TYPE_UNSPECIFIED|STRING|BOOLEAN|NUMBER|JSON; oneOf/1/$defs/artifact_json_content/properties/parameters/additionalProperties/properties/conditionalValues/additionalProperties/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; oneOf/1/$defs/artifact_json_content/properties/parameters/additionalProperties/properties/defaultValue/oneOf/0/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; oneOf/1/$defs/artifact_json_content/properties/parameters/additionalProperties/properties/valueType=|PARAMETER_VALUE_TYPE_UNSPECIFIED|STRING|BOOLEAN|NUMBER|JSON; oneOf/1/properties/encoding=json`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|batch.failed|batch.partial_success|command.canceled|command.timeout|configuration.invalid|draft.exists|expression.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|hook.failed|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|parameter.exists|profile.invalid|project.ambiguous|project.not_found|publication.cache_failed|publication.hook_failed|remote_config.conflict|remote_config.invalid|remote_config.validation_failed|resource.not_found|stdin.remote_config.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | Codes `publication.non_atomic|publication.cache_stale|publication.post_publish_hook_failed`; typed details/remediation are constrained by the shared envelope and response schema. |
| Interaction | Mode `optional`; JSON behavior `declared_conditions_return_interaction`; 2 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 3; effects `firebase_remote_read|local_cache_write|firebase_remote_write|firebase_remote_validation|local_draft_write|trusted_hook_execution|local_state_write|authentication_remote_access|local_file_write`; network `conditional` (1 conditional clauses); destructive `false`; idempotency `conditional` (8 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `add`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:add:input`; response `urn:fbrcm:schema:cli:1.0.0:command:add:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`; stdin `urn:fbrcm:schema:cli:1.0.0:stdin:remote_config`. |
| Documentation | `docs/CLI.md § fbrcm add`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `add`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/add/*_test.go`; E2E: 3 scenario(s): `parameter_add_dry_run_json`, `parameter_add_json`, `parameter_add_stateless_json`. |
| Verdict | **PASS.** |

### `auth.add.gcloud`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `auth.add.gcloud`; argv `fbrcm auth add gcloud`. |
| Arguments | `auth_id` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/auth_add_gcloud.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--label` (string, default=""); `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--quota-project` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/auth_add_gcloud.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `auth command-specific result DTO`; structural schema `schemas/cli/1.0.0/auth_add_gcloud.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/status=added`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.id_invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|profile.invalid|resource.conflict` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write|local_file_delete`; network `none` (0 conditional clauses); destructive `true`; reasons `replaces an existing auth identity and may remove its credential files`; idempotency `no` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `auth.add.gcloud`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:auth.add.gcloud:input`; response `urn:fbrcm:schema:cli:1.0.0:command:auth.add.gcloud:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm auth add gcloud`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `auth.add.gcloud`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/auth/*_test.go`; E2E: 1 scenario(s): `auth_add_gcloud_json`. |
| Verdict | **PASS.** |

### `auth.add.oauth`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `auth.add.oauth`; argv `fbrcm auth add oauth`. |
| Arguments | `auth_id` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/auth_add_oauth.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--from` (string, default=""); `--label` (string, default=""); `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--quota-project` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/auth_add_oauth.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | Modes `json_document`; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:oauth_credentials`; normalized property `schemas/cli/1.0.0/auth_add_oauth.input.schema.json#/properties/stdin`; option restrictions are schema-enforced. |
| Success | Registered DTO type(s): `auth command-specific result DTO`; structural schema `schemas/cli/1.0.0/auth_add_oauth.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/status=added`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.id_invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|interaction.required|internal.contract_violation|internal.unclassified|profile.invalid|resource.conflict|stdin.remote_config.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `missing_input_returns_interaction`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 1; effects `local_state_write|local_file_delete|local_file_write`; network `none` (0 conditional clauses); destructive `true`; reasons `replaces an existing auth identity and may remove its cached token or credential files`; idempotency `no` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `auth.add.oauth`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:auth.add.oauth:input`; response `urn:fbrcm:schema:cli:1.0.0:command:auth.add.oauth:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`; stdin `urn:fbrcm:schema:cli:1.0.0:stdin:oauth_credentials`. |
| Documentation | `docs/CLI.md § fbrcm auth add oauth`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `auth.add.oauth`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/auth/*_test.go`; E2E: 1 scenario(s): `auth_add_oauth_json`. |
| Verdict | **PASS.** |

### `auth.add.service-account`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `auth.add.service-account`; argv `fbrcm auth add service-account`. |
| Arguments | `auth_id` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/auth_add_service-account.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--from` (string, default=""); `--label` (string, default=""); `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--quota-project` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/auth_add_service-account.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | Modes `json_document`; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:service_account_credentials`; normalized property `schemas/cli/1.0.0/auth_add_service-account.input.schema.json#/properties/stdin`; option restrictions are schema-enforced. |
| Success | Registered DTO type(s): `auth command-specific result DTO`; structural schema `schemas/cli/1.0.0/auth_add_service-account.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/status=added`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.id_invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|interaction.required|internal.contract_violation|internal.unclassified|profile.invalid|resource.conflict|stdin.remote_config.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `missing_input_returns_interaction`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 1; effects `local_state_write|local_file_delete|local_file_write`; network `none` (0 conditional clauses); destructive `true`; reasons `replaces an existing auth identity and may remove its credential files`; idempotency `no` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `auth.add.service-account`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:auth.add.service-account:input`; response `urn:fbrcm:schema:cli:1.0.0:command:auth.add.service-account:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`; stdin `urn:fbrcm:schema:cli:1.0.0:stdin:service_account_credentials`. |
| Documentation | `docs/CLI.md § fbrcm auth add service-account`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `auth.add.service-account`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/auth/*_test.go`; E2E: 1 scenario(s): `auth_add_service_account_json`. |
| Verdict | **PASS.** |

### `auth.bind`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `auth.bind`; argv `fbrcm auth bind`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/auth_bind.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--auth` (string, default="", required); `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--project` (stringArray, default=[], aliases=-p, repeatable); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/auth_bind.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 2 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/auth_bind.input.schema.json`: `#/$defs/project_filter/x-fbrcm-matching`, `#/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `auth command-specific result DTO`; structural schema `schemas/cli/1.0.0/auth_bind.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/items/items/properties/status=bound|skipped`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|profile.invalid|project.ambiguous|project.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `auth.bind`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:auth.bind:input`; response `urn:fbrcm:schema:cli:1.0.0:command:auth.bind:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm auth bind`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `auth.bind`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/auth/*_test.go`; E2E: 1 scenario(s): `auth_bind_json`. |
| Verdict | **PASS.** |

### `auth.delete`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `auth.delete`; argv `fbrcm auth delete`. |
| Arguments | `auth_id` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/auth_delete.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/auth_delete.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 1 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/auth_delete.input.schema.json`: `#/properties/arguments/properties/auth_id/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `auth command-specific result DTO`; structural schema `schemas/cli/1.0.0/auth_delete.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/status=deleted; properties/type=oauth|service-account|gcloud`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.not_found|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|interaction.required|internal.contract_violation|internal.unclassified|profile.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `confirmation_required_without_bypass`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 1; effects `local_state_write|local_file_delete`; network `none` (0 conditional clauses); destructive `true`; reasons `removes stored authentication material`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `auth.delete`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:auth.delete:input`; response `urn:fbrcm:schema:cli:1.0.0:command:auth.delete:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm auth delete`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `auth.delete`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/auth/*_test.go`; E2E: 3 scenario(s): `auth_delete_gcloud_json`, `auth_delete_oauth_json`, `auth_delete_service_account_json`. |
| Verdict | **PASS.** |

### `auth.list`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `auth.list`; argv `fbrcm auth list`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/auth_list.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/auth_list.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `auth command-specific result DTO`; structural schema `schemas/cli/1.0.0/auth_list.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/items/items/properties/type=oauth|service-account|gcloud`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|profile.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `auth.list`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:auth.list:input`; response `urn:fbrcm:schema:cli:1.0.0:command:auth.list:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm auth list`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `auth.list`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/auth/*_test.go`; E2E: 1 scenario(s): `auth_list_json`. |
| Verdict | **PASS.** |

### `auth.login`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `auth.login`; argv `fbrcm auth login`. |
| Arguments | `auth_id` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/auth_login.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--noopen` (bool, default=false, ineffective); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/auth_login.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 1 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/auth_login.input.schema.json`: `#/properties/arguments/properties/auth_id/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `auth command-specific result DTO`; structural schema `schemas/cli/1.0.0/auth_login.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/paths/properties/type=oauth|service-account|gcloud; properties/status=authenticated; properties/type=oauth|service-account|gcloud`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `oauth_authorization_returns_interaction`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 2; effects `authentication_remote_access|local_file_write|local_state_write`; network `conditional` (1 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `auth.login`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:auth.login:input`; response `urn:fbrcm:schema:cli:1.0.0:command:auth.login:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm auth login`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `auth.login`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/auth/*_test.go`; E2E: none. |
| Verdict | **PASS.** |

### `auth.path`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `auth.path`; argv `fbrcm auth path`. |
| Arguments | `auth_id` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/auth_path.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/auth_path.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 1 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/auth_path.input.schema.json`: `#/properties/arguments/properties/auth_id/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `auth command-specific result DTO`; structural schema `schemas/cli/1.0.0/auth_path.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/type=oauth|service-account|gcloud`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.not_found|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|profile.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `auth.path`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:auth.path:input`; response `urn:fbrcm:schema:cli:1.0.0:command:auth.path:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm auth path`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `auth.path`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/auth/*_test.go`; E2E: 1 scenario(s): `auth_path_json`. |
| Verdict | **PASS.** |

### `auth.quota-project.set`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `auth.quota-project.set`; argv `fbrcm auth quota-project set`. |
| Arguments | `auth_id` (required; string); `quota_project_id` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/auth_quota-project_set.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/auth_quota-project_set.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `auth.quotaProjectResult`; structural schema `schemas/cli/1.0.0/auth_quota-project_set.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/source=environment|auth|credentials|unresolved; properties/status=set|unchanged`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.not_found|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|profile.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `auth.quota-project.set`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.set:input`; response `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.set:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm auth quota-project show|set|unset`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `auth.quota-project.set`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/auth/*_test.go`; E2E: 1 scenario(s): `auth_quota_project_set_json`. |
| Verdict | **PASS.** |

### `auth.quota-project.show`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `auth.quota-project.show`; argv `fbrcm auth quota-project show`. |
| Arguments | `auth_id` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/auth_quota-project_show.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/auth_quota-project_show.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `auth.quotaProjectResult`; structural schema `schemas/cli/1.0.0/auth_quota-project_show.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/source=environment|auth|credentials|unresolved`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 2; effects `authentication_remote_access|local_state_write`; network `conditional` (1 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `auth.quota-project.show`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.show:input`; response `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.show:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm auth quota-project show|set|unset`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `auth.quota-project.show`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/auth/*_test.go`; E2E: 1 scenario(s): `auth_quota_project_show_json`. |
| Verdict | **PASS.** |

### `auth.quota-project.unset`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `auth.quota-project.unset`; argv `fbrcm auth quota-project unset`. |
| Arguments | `auth_id` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/auth_quota-project_unset.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/auth_quota-project_unset.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `auth.quotaProjectResult`; structural schema `schemas/cli/1.0.0/auth_quota-project_unset.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/source=environment|auth|credentials|unresolved; properties/status=unset|unchanged`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 2; effects `local_state_write|authentication_remote_access`; network `conditional` (1 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `auth.quota-project.unset`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.unset:input`; response `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.unset:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm auth quota-project show|set|unset`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `auth.quota-project.unset`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/auth/*_test.go`; E2E: 1 scenario(s): `auth_quota_project_unset_json`. |
| Verdict | **PASS.** |

### `cache.clear`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `cache.clear`; argv `fbrcm cache clear`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/cache_clear.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/cache_clear.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `cache command-specific result DTO / shared.PathResult`; structural schema `schemas/cli/1.0.0/cache_clear.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/status=cleared|unchanged`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|interaction.required|internal.contract_violation|internal.unclassified|profile.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `confirmation_required_without_bypass`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 1; effects `local_cache_delete|local_state_write`; network `none` (0 conditional clauses); destructive `true`; reasons `removes cached Remote Config data`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `cache.clear`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:cache.clear:input`; response `urn:fbrcm:schema:cli:1.0.0:command:cache.clear:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm cache clear`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `cache.clear`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/cache/*_test.go`; E2E: 1 scenario(s): `cache_clear_json`. |
| Verdict | **PASS.** |

### `cache.list`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `cache.list`; argv `fbrcm cache list`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/cache_list.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/cache_list.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `cache command-specific result DTO / shared.PathResult`; structural schema `schemas/cli/1.0.0/cache_list.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|profile.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `cache.list`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:cache.list:input`; response `urn:fbrcm:schema:cli:1.0.0:command:cache.list:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm cache list`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `cache.list`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/cache/*_test.go`; E2E: 2 scenario(s): `cache_list_json`, `cache_list_narrow`. |
| Verdict | **PASS.** |

### `cache.path`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `cache.path`; argv `fbrcm cache path`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/cache_path.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/cache_path.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `cache command-specific result DTO / shared.PathResult`; structural schema `schemas/cli/1.0.0/cache_path.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|profile.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `cache.path`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:cache.path:input`; response `urn:fbrcm:schema:cli:1.0.0:command:cache.path:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm cache path`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `cache.path`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/cache/*_test.go`; E2E: 1 scenario(s): `cache_path_json`. |
| Verdict | **PASS.** |

### `capabilities`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `capabilities`; argv `fbrcm capabilities`. |
| Arguments | `command` (optional, repeated; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/capabilities.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", ineffective); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/capabilities.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 1 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/capabilities.input.schema.json`: `#/properties/arguments/properties/command/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `contract.CapabilityIndex | contract.Capability`; structural schema `schemas/cli/1.0.0/capabilities.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.not_executable|command.not_found|command.timeout|internal.contract_violation` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `capabilities`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:capabilities:input`; response `urn:fbrcm:schema:cli:1.0.0:command:capabilities:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm capabilities`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `capabilities`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/app/*_test.go`; E2E: 1 scenario(s): `capabilities_json`. |
| Verdict | **PASS.** |

### `completion.bash`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `completion.bash`; argv `fbrcm completion bash`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/completion_bash.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-descriptions` (bool, default=false); `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/completion_bash.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `contract.TextData`; structural schema `schemas/cli/1.0.0/completion_bash.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|internal.contract_violation` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `completion.bash`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:completion.bash:input`; response `urn:fbrcm:schema:cli:1.0.0:command:completion.bash:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm completion (shared bash/fish/powershell/zsh entry)`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `completion.bash`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/app/*_test.go`; E2E: none. |
| Verdict | **PASS.** |

### `completion.fish`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `completion.fish`; argv `fbrcm completion fish`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/completion_fish.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-descriptions` (bool, default=false); `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/completion_fish.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `contract.TextData`; structural schema `schemas/cli/1.0.0/completion_fish.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|internal.contract_violation` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `completion.fish`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:completion.fish:input`; response `urn:fbrcm:schema:cli:1.0.0:command:completion.fish:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm completion (shared bash/fish/powershell/zsh entry)`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `completion.fish`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/app/*_test.go`; E2E: none. |
| Verdict | **PASS.** |

### `completion.powershell`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `completion.powershell`; argv `fbrcm completion powershell`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/completion_powershell.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-descriptions` (bool, default=false); `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/completion_powershell.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `contract.TextData`; structural schema `schemas/cli/1.0.0/completion_powershell.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|internal.contract_violation` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `completion.powershell`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:completion.powershell:input`; response `urn:fbrcm:schema:cli:1.0.0:command:completion.powershell:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm completion (shared bash/fish/powershell/zsh entry)`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `completion.powershell`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/app/*_test.go`; E2E: none. |
| Verdict | **PASS.** |

### `completion.zsh`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `completion.zsh`; argv `fbrcm completion zsh`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/completion_zsh.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-descriptions` (bool, default=false); `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/completion_zsh.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `contract.TextData`; structural schema `schemas/cli/1.0.0/completion_zsh.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|internal.contract_violation` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `completion.zsh`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:completion.zsh:input`; response `urn:fbrcm:schema:cli:1.0.0:command:completion.zsh:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm completion (shared bash/fish/powershell/zsh entry)`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `completion.zsh`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/app/*_test.go`; E2E: none. |
| Verdict | **PASS.** |

### `conditions.add`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `conditions.add`; argv `fbrcm conditions add`. |
| Arguments | `project` (required; string); `name` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/conditions_add.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--change-note` (string, default=""); `--color` (string, default=""); `--draft` (bool, default=false, effective_when=1 clause(s)); `--dry-run` (bool, default=false); `--expression` (string, default="", required); `--no-local-config` (bool, default=false); `--priority` (int, default=0); `--profile` (string, default="", effective_when=1 clause(s)); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/conditions_add.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 1 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/conditions_add.input.schema.json`: `#/$defs/stateless_target_selector/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `[]shared/rc.RemoteMutationJSONResult`; structural schema `schemas/cli/1.0.0/conditions_add.response.schema.json#/$defs/success_data`; reachable outcome set `success|partial_success|failure`; success variants closed variants `properties/items/items/allOf/9/then/properties/error/properties/stage=publication|pre_publish_hook; properties/items/items/allOf/11/then/properties/error/properties/stage=preparation|validation|publication`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|batch.failed|batch.partial_success|command.canceled|command.timeout|condition.invalid|configuration.invalid|draft.exists|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|hook.failed|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|publication.cache_failed|publication.hook_failed|remote_config.conflict|remote_config.invalid|remote_config.validation_failed|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | Codes `publication.cache_stale|publication.post_publish_hook_failed`; typed details/remediation are constrained by the shared envelope and response schema. |
| Interaction | Mode `optional`; JSON behavior `declared_conditions_return_interaction`; 2 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 3; effects `firebase_remote_read|local_cache_write|firebase_remote_write|firebase_remote_validation|local_draft_write|trusted_hook_execution|local_state_write|authentication_remote_access|local_file_write`; network `required` (0 conditional clauses); destructive `false`; idempotency `conditional` (7 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `conditions.add`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:conditions.add:input`; response `urn:fbrcm:schema:cli:1.0.0:command:conditions.add:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § Condition mutations`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `conditions.add`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/conditions/*_test.go`; E2E: 2 scenario(s): `conditions_add_json`, `conditions_add_stateless_json`. |
| Verdict | **PASS.** |

### `conditions.delete`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `conditions.delete`; argv `fbrcm conditions delete`. |
| Arguments | `project` (required; string); `condition` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/conditions_delete.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--change-note` (string, default=""); `--draft` (bool, default=false, effective_when=1 clause(s)); `--dry-run` (bool, default=false); `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/conditions_delete.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 2 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/conditions_delete.input.schema.json`: `#/$defs/stateless_target_selector/x-fbrcm-matching`, `#/properties/arguments/properties/condition/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `[]shared/rc.RemoteMutationJSONResult`; structural schema `schemas/cli/1.0.0/conditions_delete.response.schema.json#/$defs/success_data`; reachable outcome set `success|partial_success|failure`; success variants closed variants `properties/items/items/allOf/9/then/properties/error/properties/stage=publication|pre_publish_hook; properties/items/items/allOf/11/then/properties/error/properties/stage=preparation|validation|publication`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|batch.failed|batch.partial_success|command.canceled|command.timeout|condition.invalid|condition.not_found|configuration.invalid|draft.exists|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|hook.failed|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|publication.cache_failed|publication.hook_failed|remote_config.conflict|remote_config.invalid|remote_config.validation_failed|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | Codes `publication.cache_stale|publication.post_publish_hook_failed`; typed details/remediation are constrained by the shared envelope and response schema. |
| Interaction | Mode `optional`; JSON behavior `declared_conditions_return_interaction`; 2 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 3; effects `firebase_remote_read|local_cache_write|firebase_remote_write|firebase_remote_validation|local_draft_write|trusted_hook_execution|local_state_write|authentication_remote_access|local_file_write`; network `required` (0 conditional clauses); destructive `true`; reasons `removes Remote Config conditions`; idempotency `conditional` (7 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `conditions.delete`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:conditions.delete:input`; response `urn:fbrcm:schema:cli:1.0.0:command:conditions.delete:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § Condition mutations`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `conditions.delete`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/conditions/*_test.go`; E2E: 2 scenario(s): `conditions_delete_json`, `conditions_delete_stateless_json`. |
| Verdict | **PASS.** |

### `conditions.edit`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `conditions.edit`; argv `fbrcm conditions edit`. |
| Arguments | `project` (required; string); `condition` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/conditions_edit.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--change-note` (string, default=""); `--color` (string, default=""); `--draft` (bool, default=false, effective_when=1 clause(s)); `--dry-run` (bool, default=false); `--expression` (string, default=""); `--no-color` (bool, default=false); `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/conditions_edit.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 2 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/conditions_edit.input.schema.json`: `#/$defs/stateless_target_selector/x-fbrcm-matching`, `#/properties/arguments/properties/condition/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `[]shared/rc.RemoteMutationJSONResult`; structural schema `schemas/cli/1.0.0/conditions_edit.response.schema.json#/$defs/success_data`; reachable outcome set `success|partial_success|failure`; success variants closed variants `properties/items/items/allOf/9/then/properties/error/properties/stage=publication|pre_publish_hook; properties/items/items/allOf/11/then/properties/error/properties/stage=preparation|validation|publication`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|batch.failed|batch.partial_success|command.canceled|command.timeout|condition.invalid|condition.not_found|configuration.invalid|draft.exists|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|hook.failed|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|publication.cache_failed|publication.hook_failed|remote_config.conflict|remote_config.invalid|remote_config.validation_failed|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | Codes `publication.cache_stale|publication.post_publish_hook_failed`; typed details/remediation are constrained by the shared envelope and response schema. |
| Interaction | Mode `optional`; JSON behavior `declared_conditions_return_interaction`; 2 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 3; effects `firebase_remote_read|local_cache_write|firebase_remote_write|firebase_remote_validation|local_draft_write|trusted_hook_execution|local_state_write|authentication_remote_access|local_file_write`; network `required` (0 conditional clauses); destructive `false`; idempotency `conditional` (7 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `conditions.edit`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:conditions.edit:input`; response `urn:fbrcm:schema:cli:1.0.0:command:conditions.edit:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § Condition mutations`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `conditions.edit`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/conditions/*_test.go`; E2E: 2 scenario(s): `conditions_edit_json`, `conditions_edit_stateless_json`. |
| Verdict | **PASS.** |

### `conditions.list`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `conditions.list`; argv `fbrcm conditions list`. |
| Arguments | `project` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/conditions_list.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--expr` (string, default=""); `--filter` (stringArray, default=[], aliases=-f, repeatable); `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--search` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--update` (bool, default=false, effective_when=1 clause(s)). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/conditions_list.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 4 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/conditions_list.input.schema.json`: `#/$defs/stateless_target_selector/x-fbrcm-matching`, `#/properties/options/properties/filter/items/x-fbrcm-matching`, `#/properties/options/properties/search/x-fbrcm-matching`, `#/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `[]core.ConditionEntry`; structural schema `schemas/cli/1.0.0/conditions_list.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|expression.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `oauth_authorization_returns_interaction`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 2; effects `firebase_remote_read|local_cache_write|local_state_write|authentication_remote_access|local_file_write`; network `conditional` (3 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `conditions.list`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:conditions.list:input`; response `urn:fbrcm:schema:cli:1.0.0:command:conditions.list:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm conditions list`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `conditions.list`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/conditions/*_test.go`; E2E: 2 scenario(s): `conditions_list_filtered_json`, `conditions_list_stateless_json`. |
| Verdict | **PASS.** |

### `conditions.move`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `conditions.move`; argv `fbrcm conditions move`. |
| Arguments | `project` (required; string); `condition` (required; string); `priority` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/conditions_move.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--change-note` (string, default=""); `--draft` (bool, default=false, effective_when=1 clause(s)); `--dry-run` (bool, default=false); `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/conditions_move.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 2 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/conditions_move.input.schema.json`: `#/$defs/stateless_target_selector/x-fbrcm-matching`, `#/properties/arguments/properties/condition/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `[]shared/rc.RemoteMutationJSONResult`; structural schema `schemas/cli/1.0.0/conditions_move.response.schema.json#/$defs/success_data`; reachable outcome set `success|partial_success|failure`; success variants closed variants `properties/items/items/allOf/9/then/properties/error/properties/stage=publication|pre_publish_hook; properties/items/items/allOf/11/then/properties/error/properties/stage=preparation|validation|publication`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|batch.failed|batch.partial_success|command.canceled|command.timeout|condition.invalid|condition.not_found|configuration.invalid|draft.exists|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|hook.failed|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|publication.cache_failed|publication.hook_failed|remote_config.conflict|remote_config.invalid|remote_config.validation_failed|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | Codes `publication.cache_stale|publication.post_publish_hook_failed`; typed details/remediation are constrained by the shared envelope and response schema. |
| Interaction | Mode `optional`; JSON behavior `declared_conditions_return_interaction`; 2 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 3; effects `firebase_remote_read|local_cache_write|firebase_remote_write|firebase_remote_validation|local_draft_write|trusted_hook_execution|local_state_write|authentication_remote_access|local_file_write`; network `required` (0 conditional clauses); destructive `false`; idempotency `conditional` (7 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `conditions.move`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:conditions.move:input`; response `urn:fbrcm:schema:cli:1.0.0:command:conditions.move:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § Condition mutations`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `conditions.move`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/conditions/*_test.go`; E2E: 2 scenario(s): `conditions_move_json`, `conditions_move_stateless_json`. |
| Verdict | **PASS.** |

### `conditions.rename`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `conditions.rename`; argv `fbrcm conditions rename`. |
| Arguments | `project` (required; string); `condition` (required; string); `new_name` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/conditions_rename.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--change-note` (string, default=""); `--draft` (bool, default=false, effective_when=1 clause(s)); `--dry-run` (bool, default=false); `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/conditions_rename.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 2 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/conditions_rename.input.schema.json`: `#/$defs/stateless_target_selector/x-fbrcm-matching`, `#/properties/arguments/properties/condition/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `[]shared/rc.RemoteMutationJSONResult`; structural schema `schemas/cli/1.0.0/conditions_rename.response.schema.json#/$defs/success_data`; reachable outcome set `success|partial_success|failure`; success variants closed variants `properties/items/items/allOf/9/then/properties/error/properties/stage=publication|pre_publish_hook; properties/items/items/allOf/11/then/properties/error/properties/stage=preparation|validation|publication`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|batch.failed|batch.partial_success|command.canceled|command.timeout|condition.invalid|condition.not_found|configuration.invalid|draft.exists|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|hook.failed|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|publication.cache_failed|publication.hook_failed|remote_config.conflict|remote_config.invalid|remote_config.validation_failed|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | Codes `publication.cache_stale|publication.post_publish_hook_failed`; typed details/remediation are constrained by the shared envelope and response schema. |
| Interaction | Mode `optional`; JSON behavior `declared_conditions_return_interaction`; 2 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 3; effects `firebase_remote_read|local_cache_write|firebase_remote_write|firebase_remote_validation|local_draft_write|trusted_hook_execution|local_state_write|authentication_remote_access|local_file_write`; network `required` (0 conditional clauses); destructive `false`; idempotency `conditional` (7 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `conditions.rename`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:conditions.rename:input`; response `urn:fbrcm:schema:cli:1.0.0:command:conditions.rename:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § Condition mutations`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `conditions.rename`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/conditions/*_test.go`; E2E: 2 scenario(s): `conditions_rename_json`, `conditions_rename_stateless_json`. |
| Verdict | **PASS.** |

### `conditions.show`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `conditions.show`; argv `fbrcm conditions show`. |
| Arguments | `project` (required; string); `condition` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/conditions_show.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--update` (bool, default=false, effective_when=1 clause(s)). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/conditions_show.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 2 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/conditions_show.input.schema.json`: `#/$defs/stateless_target_selector/x-fbrcm-matching`, `#/properties/arguments/properties/condition/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `conditions.conditionShowResult`; structural schema `schemas/cli/1.0.0/conditions_show.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/source=cache|cache-verified|firebase|draft`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|condition.not_found|configuration.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `oauth_authorization_returns_interaction`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 2; effects `firebase_remote_read|local_cache_write|local_state_write|authentication_remote_access|local_file_write`; network `conditional` (3 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `conditions.show`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:conditions.show:input`; response `urn:fbrcm:schema:cli:1.0.0:command:conditions.show:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm conditions show`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `conditions.show`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/conditions/*_test.go`; E2E: 2 scenario(s): `conditions_show_json`, `conditions_show_stateless_json`. |
| Verdict | **PASS.** |

### `conditions.validate`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `conditions.validate`; argv `fbrcm conditions validate`. |
| Arguments | `project` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/conditions_validate.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/conditions_validate.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 1 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/conditions_validate.input.schema.json`: `#/$defs/stateless_target_selector/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `conditions.conditionValidationResult`; structural schema `schemas/cli/1.0.0/conditions_validate.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/source=draft|firebase`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|condition.invalid|configuration.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `oauth_authorization_returns_interaction`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 2; effects `firebase_remote_read|local_cache_write|firebase_remote_validation|local_state_write|authentication_remote_access|local_file_write`; network `required` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `conditions.validate`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:conditions.validate:input`; response `urn:fbrcm:schema:cli:1.0.0:command:conditions.validate:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm conditions validate`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `conditions.validate`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/conditions/*_test.go`; E2E: 2 scenario(s): `conditions_validate_json`, `conditions_validate_stateless_json`. |
| Verdict | **PASS.** |

### `config.edit`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `config.edit`; argv `fbrcm config edit`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/config_edit.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--editor` (string, default="", ineffective); `--full` (bool, default=false, ineffective); `--no-local-config` (bool, default=false); `--profile` (string, default="", ineffective); `--scope` (string, default="global", ineffective); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/config_edit.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `explicit no-success-data registration`; structural schema `schemas/cli/1.0.0/config_edit.response.schema.json#/$defs/success_data`; reachable outcome set `failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.local_disabled|configuration.local_not_found|interaction.required|internal.contract_violation|internal.unclassified` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `required`; JSON behavior `external_input_returns_interaction`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `config.edit`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:config.edit:input`; response `urn:fbrcm:schema:cli:1.0.0:command:config.edit:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm config edit`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `config.edit`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/config/*_test.go`; E2E: none. |
| Verdict | **PASS.** |

### `config.path`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `config.path`; argv `fbrcm config path`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/config_path.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", ineffective); `--scope` (string, default="global"); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/config_path.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `config command-specific result DTO`; structural schema `schemas/cli/1.0.0/config_path.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/scope=global|local`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|configuration.local_disabled|configuration.local_not_found|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `config.path`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:config.path:input`; response `urn:fbrcm:schema:cli:1.0.0:command:config.path:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm config path`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `config.path`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/config/*_test.go`; E2E: 1 scenario(s): `config_path_json`. |
| Verdict | **PASS.** |

### `config.reset`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `config.reset`; argv `fbrcm config reset`. |
| Arguments | `key` (optional; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/config_reset.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", ineffective); `--scope` (string, default="global"); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/config_reset.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `config command-specific result DTO`; structural schema `schemas/cli/1.0.0/config_reset.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/scope=global|local; properties/status=unchanged|reset`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|configuration.local_disabled|configuration.local_not_found|file.io_failed|filesystem.permission_denied|interaction.required|internal.contract_violation|internal.unclassified` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `confirmation_required_without_bypass`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `true`; reasons `removes persisted configuration values`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `config.reset`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:config.reset:input`; response `urn:fbrcm:schema:cli:1.0.0:command:config.reset:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm config reset`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `config.reset`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/config/*_test.go`; E2E: 1 scenario(s): `config_reset_json`. |
| Verdict | **PASS.** |

### `config.set`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `config.set`; argv `fbrcm config set`. |
| Arguments | `key` (required; string); `value` (required, repeated; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/config_set.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", ineffective); `--scope` (string, default="global"); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/config_set.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `config command-specific result DTO`; structural schema `schemas/cli/1.0.0/config_set.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/scope=global|local`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|configuration.local_disabled|configuration.local_not_found|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `config.set`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:config.set:input`; response `urn:fbrcm:schema:cli:1.0.0:command:config.set:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm config set`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `config.set`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/config/*_test.go`; E2E: 2 scenario(s): `config_set_json`, `config_set_theme_json`. |
| Verdict | **PASS.** |

### `config.show`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `config.show`; argv `fbrcm config show`. |
| Arguments | `key` (optional; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/config_show.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", ineffective); `--scope` (string, default="effective"); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/config_show.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `config command-specific result DTO`; structural schema `schemas/cli/1.0.0/config_show.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `oneOf/0/properties/scope=effective|global|local; oneOf/1/properties/source=absent|default|global|local|mixed|migrated`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|configuration.local_disabled|configuration.local_not_found|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `config.show`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:config.show:input`; response `urn:fbrcm:schema:cli:1.0.0:command:config.show:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm config show`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `config.show`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/config/*_test.go`; E2E: 1 scenario(s): `config_show_json`. |
| Verdict | **PASS.** |

### `config.validate`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `config.validate`; argv `fbrcm config validate`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/config_validate.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", ineffective); `--scope` (string, default="all"); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/config_validate.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `config command-specific result DTO`; structural schema `schemas/cli/1.0.0/config_validate.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/errors/items/properties/code=duplicate_binding|empty_binding|invalid_binding|invalid_profile|keybinding_conflict|legacy_bindings|missing_profile|project_alias_source|repository_scope_required|toml_decode|unknown_action|unknown_block; properties/warnings/items/properties/code=duplicate_binding|empty_binding|invalid_binding|invalid_profile|keybinding_conflict|legacy_bindings|missing_profile|project_alias_source|repository_scope_required|toml_decode|unknown_action|unknown_block`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|configuration.local_disabled|configuration.local_not_found|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|validation.failed` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `config.validate`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:config.validate:input`; response `urn:fbrcm:schema:cli:1.0.0:command:config.validate:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm config validate`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `config.validate`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/config/*_test.go`; E2E: 2 scenario(s): `config_validate_invalid_json`, `config_validate_valid_json`. |
| Verdict | **PASS.** |

### `delete`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `delete`; argv `fbrcm delete`. |
| Arguments | `parameter` (optional; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/delete.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--change-note` (string, default=""); `--draft` (bool, default=false, effective_when=1 clause(s)); `--dry-run` (bool, default=false); `--expr` (string, default=""); `--filter` (stringArray, default=[], aliases=-f, repeatable); `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--project` (stringArray, default=[], aliases=-p, repeatable); `--search` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/delete.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 6 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/delete.input.schema.json`: `#/$defs/stateless_get_project_selector/oneOf/0/x-fbrcm-matching`, `#/$defs/stateless_get_project_selector/oneOf/1/x-fbrcm-matching`, `#/properties/arguments/properties/parameter/x-fbrcm-matching`, `#/properties/options/properties/filter/items/x-fbrcm-matching`, `#/properties/options/properties/search/x-fbrcm-matching`, `#/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | Modes `json_document`; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:remote_config`; normalized property `schemas/cli/1.0.0/delete.input.schema.json#/properties/stdin`; option restrictions are schema-enforced. |
| Success | Registered DTO type(s): `[]shared/rc.RemoteMutationJSONResult | contract.ArtifactData`; structural schema `schemas/cli/1.0.0/delete.response.schema.json#/$defs/success_data`; reachable outcome set `success|partial_success|failure`; success variants closed variants `oneOf/0/properties/items/items/allOf/9/then/properties/error/properties/stage=publication|pre_publish_hook; oneOf/0/properties/items/items/allOf/11/then/properties/error/properties/stage=preparation|validation|publication; oneOf/1/$defs/artifact_json_content/properties/parameterGroups/additionalProperties/properties/parameters/additionalProperties/properties/conditionalValues/additionalProperties/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; oneOf/1/$defs/artifact_json_content/properties/parameterGroups/additionalProperties/properties/parameters/additionalProperties/properties/defaultValue/oneOf/0/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; oneOf/1/$defs/artifact_json_content/properties/parameterGroups/additionalProperties/properties/parameters/additionalProperties/properties/valueType=|PARAMETER_VALUE_TYPE_UNSPECIFIED|STRING|BOOLEAN|NUMBER|JSON; oneOf/1/$defs/artifact_json_content/properties/parameters/additionalProperties/properties/conditionalValues/additionalProperties/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; oneOf/1/$defs/artifact_json_content/properties/parameters/additionalProperties/properties/defaultValue/oneOf/0/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; oneOf/1/$defs/artifact_json_content/properties/parameters/additionalProperties/properties/valueType=|PARAMETER_VALUE_TYPE_UNSPECIFIED|STRING|BOOLEAN|NUMBER|JSON; oneOf/1/properties/encoding=json`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|batch.failed|batch.partial_success|command.canceled|command.timeout|configuration.invalid|draft.exists|expression.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|hook.failed|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|parameter.ambiguous|parameter.not_found|profile.invalid|project.ambiguous|project.not_found|publication.cache_failed|publication.hook_failed|remote_config.conflict|remote_config.invalid|remote_config.validation_failed|resource.not_found|stdin.remote_config.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | Codes `publication.non_atomic|publication.cache_stale|publication.post_publish_hook_failed`; typed details/remediation are constrained by the shared envelope and response schema. |
| Interaction | Mode `optional`; JSON behavior `declared_conditions_return_interaction`; 2 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 3; effects `firebase_remote_read|local_cache_write|firebase_remote_write|firebase_remote_validation|local_draft_write|trusted_hook_execution|local_state_write|authentication_remote_access|local_file_write`; network `conditional` (1 conditional clauses); destructive `true`; reasons `removes Remote Config parameters`; idempotency `conditional` (8 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `delete`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:delete:input`; response `urn:fbrcm:schema:cli:1.0.0:command:delete:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`; stdin `urn:fbrcm:schema:cli:1.0.0:stdin:remote_config`. |
| Documentation | `docs/CLI.md § fbrcm delete`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `delete`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/delete/*_test.go`; E2E: 4 scenario(s): `parameter_delete_dry_run_json`, `parameter_delete_json`, `parameter_delete_stateless_json`, `project_import_stateless_cleanup_json`. |
| Verdict | **PASS.** |

### `doctor`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `doctor`; argv `fbrcm doctor`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/doctor.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/doctor.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `[]doctor.doctorListItem`; structural schema `schemas/cli/1.0.0/doctor.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/items/items/properties/status=pass|warn|fail`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|diagnostic.failed|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 2; effects `local_file_write|local_file_delete|firebase_remote_read|authentication_remote_access|local_state_write`; network `conditional` (1 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `doctor`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:doctor:input`; response `urn:fbrcm:schema:cli:1.0.0:command:doctor:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm doctor`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `doctor`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/doctor/*_test.go`; E2E: 1 scenario(s): `doctor_json`. |
| Verdict | **PASS.** |

### `draft.change-note`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `draft.change-note`; argv `fbrcm draft change-note`. |
| Arguments | `project` (required; string); `text` (optional; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/draft_change-note.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--clear` (bool, default=false); `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/draft_change-note.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `draft command-specific DTO / contract.ArtifactData / shared.PathResult`; structural schema `schemas/cli/1.0.0/draft_change-note.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|draft.ambiguous|draft.not_found|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|profile.invalid|project.ambiguous|project.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_draft_write|local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `conditional` (3 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `draft.change-note`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:draft.change-note:input`; response `urn:fbrcm:schema:cli:1.0.0:command:draft.change-note:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm draft change-note`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `draft.change-note`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/draft/*_test.go`; E2E: 1 scenario(s): `draft_change_note_json`. |
| Verdict | **PASS.** |

### `draft.diff`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `draft.diff`; argv `fbrcm draft diff`. |
| Arguments | `project` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/draft_diff.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--against` (string, default="base"); `--cached` (bool, default=false); `--conditions` (bool, default=false); `--expr` (string, default=""); `--filter` (stringArray, default=[], aliases=-f, repeatable); `--group` (stringArray, default=[], repeatable); `--no-local-config` (bool, default=false); `--parameters` (bool, default=false); `--profile` (string, default=""); `--search` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/draft_diff.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 3 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/draft_diff.input.schema.json`: `#/properties/options/properties/filter/items/x-fbrcm-matching`, `#/properties/options/properties/search/x-fbrcm-matching`, `#/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `draft command-specific DTO / contract.ArtifactData / shared.PathResult`; structural schema `schemas/cli/1.0.0/draft_diff.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/against=base|current; properties/diff/properties/conditions/items/properties/kind=added|removed|changed|unchanged; properties/diff/properties/group_descriptions/items/properties/kind=added|removed|changed|unchanged; properties/diff/properties/parameters/items/properties/current/properties/conditionalValues/additionalProperties/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; properties/diff/properties/parameters/items/properties/current/properties/defaultValue/oneOf/0/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; properties/diff/properties/parameters/items/properties/current/properties/valueType=|PARAMETER_VALUE_TYPE_UNSPECIFIED|STRING|BOOLEAN|NUMBER|JSON; properties/diff/properties/parameters/items/properties/final/properties/conditionalValues/additionalProperties/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; properties/diff/properties/parameters/items/properties/final/properties/defaultValue/oneOf/0/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; properties/diff/properties/parameters/items/properties/final/properties/valueType=|PARAMETER_VALUE_TYPE_UNSPECIFIED|STRING|BOOLEAN|NUMBER|JSON; properties/diff/properties/parameters/items/properties/kind=added|removed|changed|unchanged`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|draft.ambiguous|draft.not_found|expression.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `oauth_authorization_returns_interaction`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 2; effects `firebase_remote_read|local_cache_write|authentication_remote_access|local_file_write|local_state_write`; network `conditional` (1 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `draft.diff`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:draft.diff:input`; response `urn:fbrcm:schema:cli:1.0.0:command:draft.diff:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm draft diff`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `draft.diff`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/draft/*_test.go`; E2E: 1 scenario(s): `draft_diff_json`. |
| Verdict | **PASS.** |

### `draft.discard`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `draft.discard`; argv `fbrcm draft discard`. |
| Arguments | `project` (optional, repeated; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/draft_discard.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--all` (bool, default=false); `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/draft_discard.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 1 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/draft_discard.input.schema.json`: `#/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `draft command-specific DTO / contract.ArtifactData / shared.PathResult`; structural schema `schemas/cli/1.0.0/draft_discard.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/items/items/properties/status=discarded`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|draft.ambiguous|draft.not_found|file.io_failed|filesystem.permission_denied|interaction.required|internal.contract_violation|internal.unclassified|profile.invalid|project.ambiguous|project.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `confirmation_required_without_bypass`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 1; effects `local_draft_delete|local_state_write`; network `none` (0 conditional clauses); destructive `true`; reasons `removes local Remote Config drafts`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `draft.discard`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:draft.discard:input`; response `urn:fbrcm:schema:cli:1.0.0:command:draft.discard:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm draft discard`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `draft.discard`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/draft/*_test.go`; E2E: 1 scenario(s): `draft_discard_json`. |
| Verdict | **PASS.** |

### `draft.list`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `draft.list`; argv `fbrcm draft list`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/draft_list.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--filter` (stringArray, default=[], aliases=-f, repeatable); `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/draft_list.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 1 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/draft_list.input.schema.json`: `#/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `draft command-specific DTO / contract.ArtifactData / shared.PathResult`; structural schema `schemas/cli/1.0.0/draft_list.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/items/items/properties/status=invalid|ready|unchanged`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|profile.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `draft.list`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:draft.list:input`; response `urn:fbrcm:schema:cli:1.0.0:command:draft.list:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm draft list`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `draft.list`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/draft/*_test.go`; E2E: 2 scenario(s): `draft_list_json`, `draft_list_narrow`. |
| Verdict | **PASS.** |

### `draft.path`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `draft.path`; argv `fbrcm draft path`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/draft_path.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/draft_path.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `draft command-specific DTO / contract.ArtifactData / shared.PathResult`; structural schema `schemas/cli/1.0.0/draft_path.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|profile.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `draft.path`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:draft.path:input`; response `urn:fbrcm:schema:cli:1.0.0:command:draft.path:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm draft path`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `draft.path`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/draft/*_test.go`; E2E: 1 scenario(s): `draft_path_json`. |
| Verdict | **PASS.** |

### `draft.publish`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `draft.publish`; argv `fbrcm draft publish`. |
| Arguments | `project` (optional, repeated; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/draft_publish.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--all` (bool, default=false); `--change-note` (string, default=""); `--dry-run` (bool, default=false); `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/draft_publish.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 1 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/draft_publish.input.schema.json`: `#/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `draft command-specific DTO / contract.ArtifactData / shared.PathResult`; structural schema `schemas/cli/1.0.0/draft_publish.response.schema.json#/$defs/success_data`; reachable outcome set `success|partial_success|failure`; success variants closed variants `properties/items/items/properties/status=failed|unchanged|would-publish|already-applied|published|published-hook-failed|published-cache-failed|published-cleanup-failed|conflict`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|batch.failed|batch.partial_success|command.canceled|command.timeout|configuration.invalid|draft.ambiguous|draft.not_found|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|hook.failed|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|parameters_cache.not_found|profile.invalid|project.ambiguous|project.not_found|publication.cache_failed|publication.hook_failed|remote_config.conflict|remote_config.invalid|remote_config.validation_failed|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | Codes `publication.cache_stale|publication.draft_cleanup_failed|publication.non_atomic|publication.post_publish_hook_failed`; typed details/remediation are constrained by the shared envelope and response schema. |
| Interaction | Mode `optional`; JSON behavior `declared_conditions_return_interaction`; 2 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 3; effects `firebase_remote_read|local_cache_write|firebase_remote_write|firebase_remote_validation|trusted_hook_execution|local_draft_delete|local_draft_write|authentication_remote_access|local_file_write|local_state_write`; network `required` (0 conditional clauses); destructive `true`; reasons `a draft may replace or remove current Remote Config content`; idempotency `conditional` (6 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `draft.publish`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:draft.publish:input`; response `urn:fbrcm:schema:cli:1.0.0:command:draft.publish:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm draft publish`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `draft.publish`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/draft/*_test.go`; E2E: none. |
| Verdict | **PASS.** |

### `draft.show`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `draft.show`; argv `fbrcm draft show`. |
| Arguments | `project` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/draft_show.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--raw` (bool, default=false); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--to` (string, default=""). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/draft_show.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `draft command-specific DTO / contract.ArtifactData / shared.PathResult`; structural schema `schemas/cli/1.0.0/draft_show.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/encoding=json|utf-8|base64|none`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|draft.ambiguous|draft.not_found|file.io_failed|filesystem.permission_denied|interaction.required|internal.contract_violation|internal.unclassified|profile.invalid|project.ambiguous|project.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `destination_conflict_returns_interaction`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 1; effects `local_file_write|local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `draft.show`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:draft.show:input`; response `urn:fbrcm:schema:cli:1.0.0:command:draft.show:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm draft show`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `draft.show`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/draft/*_test.go`; E2E: 1 scenario(s): `draft_show_json`. |
| Verdict | **PASS.** |

### `duplicate`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `duplicate`; argv `fbrcm duplicate`. |
| Arguments | `source` (required; string); `target` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/duplicate.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--change-note` (string, default=""); `--draft` (bool, default=false, effective_when=1 clause(s)); `--dry-run` (bool, default=false); `--expr` (string, default=""); `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--project` (stringArray, default=[], aliases=-p, repeatable); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/duplicate.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 4 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/duplicate.input.schema.json`: `#/$defs/stateless_get_project_selector/oneOf/0/x-fbrcm-matching`, `#/$defs/stateless_get_project_selector/oneOf/1/x-fbrcm-matching`, `#/properties/arguments/properties/source/x-fbrcm-matching`, `#/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `[]shared/rc.RemoteMutationJSONResult`; structural schema `schemas/cli/1.0.0/duplicate.response.schema.json#/$defs/success_data`; reachable outcome set `success|partial_success|failure`; success variants closed variants `properties/items/items/allOf/9/then/properties/error/properties/stage=publication|pre_publish_hook; properties/items/items/allOf/11/then/properties/error/properties/stage=preparation|validation|publication`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|batch.failed|batch.partial_success|command.canceled|command.timeout|configuration.invalid|draft.exists|expression.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|hook.failed|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|parameter.ambiguous|parameter.exists|parameter.not_found|profile.invalid|project.ambiguous|project.not_found|publication.cache_failed|publication.hook_failed|remote_config.conflict|remote_config.invalid|remote_config.validation_failed|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | Codes `publication.non_atomic|publication.cache_stale|publication.post_publish_hook_failed`; typed details/remediation are constrained by the shared envelope and response schema. |
| Interaction | Mode `optional`; JSON behavior `declared_conditions_return_interaction`; 2 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 3; effects `firebase_remote_read|local_cache_write|firebase_remote_write|firebase_remote_validation|local_draft_write|trusted_hook_execution|local_state_write|authentication_remote_access|local_file_write`; network `required` (0 conditional clauses); destructive `false`; idempotency `conditional` (7 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `duplicate`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:duplicate:input`; response `urn:fbrcm:schema:cli:1.0.0:command:duplicate:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm duplicate`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `duplicate`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/duplicate/*_test.go`; E2E: 3 scenario(s): `parameter_duplicate_dry_run_json`, `parameter_duplicate_json`, `parameter_duplicate_stateless_json`. |
| Verdict | **PASS.** |

### `experiments.delete`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `experiments.delete`; argv `fbrcm experiments delete`. |
| Arguments | `project` (required; string); `experiment_id` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/experiments_delete.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/experiments_delete.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `managedfeatures command-specific DTO / core entry DTO`; structural schema `schemas/cli/1.0.0/experiments_delete.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/status=deleted|would-delete`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `declared_conditions_return_interaction`; 2 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 3; effects `firebase_remote_read|firebase_managed_feature_delete|local_state_write|authentication_remote_access|local_file_write`; network `required` (0 conditional clauses); destructive `true`; reasons `removes a Firebase Remote Config experiment`; idempotency `conditional` (3 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `experiments.delete`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:experiments.delete:input`; response `urn:fbrcm:schema:cli:1.0.0:command:experiments.delete:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm experiments delete`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `experiments.delete`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/managedfeatures/*_test.go`; E2E: 1 scenario(s): `experiments_delete_stateless_confirmation_json`. |
| Verdict | **PASS.** |

### `experiments.list`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `experiments.list`; argv `fbrcm experiments list`. |
| Arguments | `project` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/experiments_list.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--filter` (stringArray, default=[], aliases=-f, repeatable); `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--update` (bool, default=false, effective_when=1 clause(s)). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/experiments_list.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 2 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/experiments_list.input.schema.json`: `#/properties/options/properties/filter/items/x-fbrcm-matching`, `#/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `managedfeatures command-specific DTO / core entry DTO`; structural schema `schemas/cli/1.0.0/experiments_list.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `oauth_authorization_returns_interaction`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 2; effects `firebase_remote_read|local_cache_write|local_state_write|authentication_remote_access|local_file_write`; network `required` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `experiments.list`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:experiments.list:input`; response `urn:fbrcm:schema:cli:1.0.0:command:experiments.list:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm experiments list`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `experiments.list`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/managedfeatures/*_test.go`; E2E: 2 scenario(s): `experiments_list_json`, `experiments_list_stateless_json`. |
| Verdict | **PASS.** |

### `experiments.show`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `experiments.show`; argv `fbrcm experiments show`. |
| Arguments | `project` (required; string); `experiment_id` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/experiments_show.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--update` (bool, default=false, effective_when=1 clause(s)). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/experiments_show.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `managedfeatures command-specific DTO / core entry DTO`; structural schema `schemas/cli/1.0.0/experiments_show.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `oauth_authorization_returns_interaction`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 2; effects `firebase_remote_read|local_cache_write|local_state_write|authentication_remote_access|local_file_write`; network `required` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `experiments.show`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:experiments.show:input`; response `urn:fbrcm:schema:cli:1.0.0:command:experiments.show:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm experiments show`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `experiments.show`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/managedfeatures/*_test.go`; E2E: 2 scenario(s): `experiments_show_json`, `experiments_show_stateless_json`. |
| Verdict | **PASS.** |

### `get`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `get`; argv `fbrcm get`. |
| Arguments | `parameter` (optional; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/get.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--all` (bool, default=false); `--expr` (string, default=""); `--filter` (stringArray, default=[], aliases=-f, repeatable); `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--project` (stringArray, default=[], aliases=-p, repeatable); `--search` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--update` (bool, default=false, effective_when=1 clause(s)). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/get.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 6 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/get.input.schema.json`: `#/$defs/stateless_get_project_selector/oneOf/0/x-fbrcm-matching`, `#/$defs/stateless_get_project_selector/oneOf/1/x-fbrcm-matching`, `#/properties/arguments/properties/parameter/x-fbrcm-matching`, `#/properties/options/properties/filter/items/x-fbrcm-matching`, `#/properties/options/properties/search/x-fbrcm-matching`, `#/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | Modes `json_document`; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:remote_config`; normalized property `schemas/cli/1.0.0/get.input.schema.json#/properties/stdin`; option restrictions are schema-enforced. |
| Success | Registered DTO type(s): `[]get.parameterRowJSON`; structural schema `schemas/cli/1.0.0/get.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/items/items/properties/status=fetch|cached|stale|missing|error`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|expression.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|parameter.ambiguous|parameter.not_found|profile.invalid|project.ambiguous|project.not_found|resource.not_found|stdin.remote_config.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | Codes `cache.stale`; typed details/remediation are constrained by the shared envelope and response schema. |
| Interaction | Mode `optional`; JSON behavior `oauth_authorization_returns_interaction`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 2; effects `firebase_remote_read|local_cache_write|local_state_write|authentication_remote_access|local_file_write`; network `conditional` (3 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `get`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:get:input`; response `urn:fbrcm:schema:cli:1.0.0:command:get:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`; stdin `urn:fbrcm:schema:cli:1.0.0:stdin:remote_config`. |
| Documentation | `docs/CLI.md § fbrcm get`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `get`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/get/*_test.go`; E2E: 8 scenario(s): `get_filtered_json`, `get_stateless_all_projects_json`, `get_stateless_contains_project_json`, `get_stateless_exact_project_json`, `get_stateless_explicit_fuzzy_project_json`, `get_stateless_json`, `get_stateless_starts_with_project_json`, `get_update`. |
| Verdict | **PASS.** |

### `groups.add`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `groups.add`; argv `fbrcm groups add`. |
| Arguments | `name` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/groups_add.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--change-note` (string, default=""); `--description` (string, default=""); `--draft` (bool, default=false, effective_when=1 clause(s)); `--dry-run` (bool, default=false); `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--project` (stringArray, default=[], aliases=-p, repeatable); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/groups_add.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 3 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/groups_add.input.schema.json`: `#/$defs/stateless_get_project_selector/oneOf/0/x-fbrcm-matching`, `#/$defs/stateless_get_project_selector/oneOf/1/x-fbrcm-matching`, `#/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `[]shared/rc.RemoteMutationJSONResult`; structural schema `schemas/cli/1.0.0/groups_add.response.schema.json#/$defs/success_data`; reachable outcome set `success|partial_success|failure`; success variants closed variants `properties/items/items/allOf/9/then/properties/error/properties/stage=publication|pre_publish_hook; properties/items/items/allOf/11/then/properties/error/properties/stage=preparation|validation|publication`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|batch.failed|batch.partial_success|command.canceled|command.timeout|configuration.invalid|draft.exists|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|group.not_found|hook.failed|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|publication.cache_failed|publication.hook_failed|remote_config.conflict|remote_config.invalid|remote_config.validation_failed|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | Codes `publication.non_atomic|publication.cache_stale|publication.post_publish_hook_failed`; typed details/remediation are constrained by the shared envelope and response schema. |
| Interaction | Mode `optional`; JSON behavior `declared_conditions_return_interaction`; 2 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 3; effects `firebase_remote_read|local_cache_write|firebase_remote_write|firebase_remote_validation|local_draft_write|trusted_hook_execution|local_state_write|authentication_remote_access|local_file_write`; network `required` (0 conditional clauses); destructive `false`; idempotency `conditional` (7 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `groups.add`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:groups.add:input`; response `urn:fbrcm:schema:cli:1.0.0:command:groups.add:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § Group mutations`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `groups.add`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/groups/*_test.go`; E2E: 2 scenario(s): `groups_add_json`, `groups_add_stateless_json`. |
| Verdict | **PASS.** |

### `groups.delete`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `groups.delete`; argv `fbrcm groups delete`. |
| Arguments | `group` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/groups_delete.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--change-note` (string, default=""); `--draft` (bool, default=false, effective_when=1 clause(s)); `--dry-run` (bool, default=false); `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--project` (stringArray, default=[], aliases=-p, repeatable); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/groups_delete.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 4 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/groups_delete.input.schema.json`: `#/$defs/stateless_get_project_selector/oneOf/0/x-fbrcm-matching`, `#/$defs/stateless_get_project_selector/oneOf/1/x-fbrcm-matching`, `#/properties/arguments/properties/group/x-fbrcm-matching`, `#/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `[]shared/rc.RemoteMutationJSONResult`; structural schema `schemas/cli/1.0.0/groups_delete.response.schema.json#/$defs/success_data`; reachable outcome set `success|partial_success|failure`; success variants closed variants `properties/items/items/allOf/9/then/properties/error/properties/stage=publication|pre_publish_hook; properties/items/items/allOf/11/then/properties/error/properties/stage=preparation|validation|publication`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|batch.failed|batch.partial_success|command.canceled|command.timeout|configuration.invalid|draft.exists|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|group.not_found|hook.failed|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|publication.cache_failed|publication.hook_failed|remote_config.conflict|remote_config.invalid|remote_config.validation_failed|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | Codes `publication.non_atomic|publication.cache_stale|publication.post_publish_hook_failed`; typed details/remediation are constrained by the shared envelope and response schema. |
| Interaction | Mode `optional`; JSON behavior `declared_conditions_return_interaction`; 2 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 3; effects `firebase_remote_read|local_cache_write|firebase_remote_write|firebase_remote_validation|local_draft_write|trusted_hook_execution|local_state_write|authentication_remote_access|local_file_write`; network `required` (0 conditional clauses); destructive `true`; reasons `removes Remote Config parameter groups`; idempotency `conditional` (7 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `groups.delete`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:groups.delete:input`; response `urn:fbrcm:schema:cli:1.0.0:command:groups.delete:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § Group mutations`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `groups.delete`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/groups/*_test.go`; E2E: 2 scenario(s): `groups_delete_json`, `groups_delete_stateless_json`. |
| Verdict | **PASS.** |

### `groups.edit`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `groups.edit`; argv `fbrcm groups edit`. |
| Arguments | `group` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/groups_edit.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--change-note` (string, default=""); `--description` (string, default=""); `--draft` (bool, default=false, effective_when=1 clause(s)); `--dry-run` (bool, default=false); `--no-description` (bool, default=false); `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--project` (stringArray, default=[], aliases=-p, repeatable); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/groups_edit.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 4 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/groups_edit.input.schema.json`: `#/$defs/stateless_get_project_selector/oneOf/0/x-fbrcm-matching`, `#/$defs/stateless_get_project_selector/oneOf/1/x-fbrcm-matching`, `#/properties/arguments/properties/group/x-fbrcm-matching`, `#/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `[]shared/rc.RemoteMutationJSONResult`; structural schema `schemas/cli/1.0.0/groups_edit.response.schema.json#/$defs/success_data`; reachable outcome set `success|partial_success|failure`; success variants closed variants `properties/items/items/allOf/9/then/properties/error/properties/stage=publication|pre_publish_hook; properties/items/items/allOf/11/then/properties/error/properties/stage=preparation|validation|publication`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|batch.failed|batch.partial_success|command.canceled|command.timeout|configuration.invalid|draft.exists|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|group.not_found|hook.failed|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|publication.cache_failed|publication.hook_failed|remote_config.conflict|remote_config.invalid|remote_config.validation_failed|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | Codes `publication.non_atomic|publication.cache_stale|publication.post_publish_hook_failed`; typed details/remediation are constrained by the shared envelope and response schema. |
| Interaction | Mode `optional`; JSON behavior `declared_conditions_return_interaction`; 2 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 3; effects `firebase_remote_read|local_cache_write|firebase_remote_write|firebase_remote_validation|local_draft_write|trusted_hook_execution|local_state_write|authentication_remote_access|local_file_write`; network `required` (0 conditional clauses); destructive `false`; idempotency `conditional` (7 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `groups.edit`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:groups.edit:input`; response `urn:fbrcm:schema:cli:1.0.0:command:groups.edit:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § Group mutations`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `groups.edit`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/groups/*_test.go`; E2E: 2 scenario(s): `groups_edit_json`, `groups_edit_stateless_json`. |
| Verdict | **PASS.** |

### `groups.list`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `groups.list`; argv `fbrcm groups list`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/groups_list.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--filter` (stringArray, default=[], aliases=-f, repeatable); `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--project` (stringArray, default=[], aliases=-p, repeatable); `--search` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--update` (bool, default=false, effective_when=1 clause(s)). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/groups_list.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 5 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/groups_list.input.schema.json`: `#/$defs/stateless_get_project_selector/oneOf/0/x-fbrcm-matching`, `#/$defs/stateless_get_project_selector/oneOf/1/x-fbrcm-matching`, `#/properties/options/properties/filter/items/x-fbrcm-matching`, `#/properties/options/properties/search/x-fbrcm-matching`, `#/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `[]groups.groupListJSON`; structural schema `schemas/cli/1.0.0/groups_list.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/items/items/properties/source=cache|cache-verified|firebase|draft`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `oauth_authorization_returns_interaction`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 2; effects `firebase_remote_read|local_cache_write|local_state_write|authentication_remote_access|local_file_write`; network `conditional` (3 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `groups.list`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:groups.list:input`; response `urn:fbrcm:schema:cli:1.0.0:command:groups.list:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm groups list`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `groups.list`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/groups/*_test.go`; E2E: 2 scenario(s): `groups_list_filtered_json`, `groups_list_stateless_json`. |
| Verdict | **PASS.** |

### `groups.rename`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `groups.rename`; argv `fbrcm groups rename`. |
| Arguments | `group` (required; string); `new_name` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/groups_rename.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--change-note` (string, default=""); `--draft` (bool, default=false, effective_when=1 clause(s)); `--dry-run` (bool, default=false); `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--project` (stringArray, default=[], aliases=-p, repeatable); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/groups_rename.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 4 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/groups_rename.input.schema.json`: `#/$defs/stateless_get_project_selector/oneOf/0/x-fbrcm-matching`, `#/$defs/stateless_get_project_selector/oneOf/1/x-fbrcm-matching`, `#/properties/arguments/properties/group/x-fbrcm-matching`, `#/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `[]shared/rc.RemoteMutationJSONResult`; structural schema `schemas/cli/1.0.0/groups_rename.response.schema.json#/$defs/success_data`; reachable outcome set `success|partial_success|failure`; success variants closed variants `properties/items/items/allOf/9/then/properties/error/properties/stage=publication|pre_publish_hook; properties/items/items/allOf/11/then/properties/error/properties/stage=preparation|validation|publication`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|batch.failed|batch.partial_success|command.canceled|command.timeout|configuration.invalid|draft.exists|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|group.not_found|hook.failed|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|publication.cache_failed|publication.hook_failed|remote_config.conflict|remote_config.invalid|remote_config.validation_failed|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | Codes `publication.non_atomic|publication.cache_stale|publication.post_publish_hook_failed`; typed details/remediation are constrained by the shared envelope and response schema. |
| Interaction | Mode `optional`; JSON behavior `declared_conditions_return_interaction`; 2 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 3; effects `firebase_remote_read|local_cache_write|firebase_remote_write|firebase_remote_validation|local_draft_write|trusted_hook_execution|local_state_write|authentication_remote_access|local_file_write`; network `required` (0 conditional clauses); destructive `false`; idempotency `conditional` (7 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `groups.rename`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:groups.rename:input`; response `urn:fbrcm:schema:cli:1.0.0:command:groups.rename:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § Group mutations`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `groups.rename`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/groups/*_test.go`; E2E: 2 scenario(s): `groups_rename_json`, `groups_rename_stateless_json`. |
| Verdict | **PASS.** |

### `help`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `help`; argv `fbrcm help`. |
| Arguments | `command` (optional, repeated; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/help.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", ineffective); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/help.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 1 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/help.input.schema.json`: `#/properties/arguments/properties/command/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `contract.TextData`; structural schema `schemas/cli/1.0.0/help.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|internal.contract_violation` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `help`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:help:input`; response `urn:fbrcm:schema:cli:1.0.0:command:help:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm help`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `help`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/app/*_test.go`; E2E: 2 scenario(s): `help_projects_diff`, `help_projects_diff_json`. |
| Verdict | **PASS.** |

### `hooks.fingerprint`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `hooks.fingerprint`; argv `fbrcm hooks fingerprint`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/hooks_fingerprint.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", ineffective); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/hooks_fingerprint.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `hooks command-specific result DTO`; structural schema `schemas/cli/1.0.0/hooks_fingerprint.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|hooks.not_configured|internal.contract_violation|internal.unclassified` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `hooks.fingerprint`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:hooks.fingerprint:input`; response `urn:fbrcm:schema:cli:1.0.0:command:hooks.fingerprint:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm hooks fingerprint`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `hooks.fingerprint`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/hooks/*_test.go`; E2E: 1 scenario(s): `hooks_fingerprint_json`. |
| Verdict | **PASS.** |

### `hooks.status`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `hooks.status`; argv `fbrcm hooks status`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/hooks_status.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", ineffective); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/hooks_status.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `hooks command-specific result DTO`; structural schema `schemas/cli/1.0.0/hooks_status.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `hooks.status`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:hooks.status:input`; response `urn:fbrcm:schema:cli:1.0.0:command:hooks.status:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm hooks status`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `hooks.status`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/hooks/*_test.go`; E2E: 1 scenario(s): `hooks_status_json`. |
| Verdict | **PASS.** |

### `hooks.trust`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `hooks.trust`; argv `fbrcm hooks trust`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/hooks_trust.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", ineffective); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/hooks_trust.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `hooks command-specific result DTO`; structural schema `schemas/cli/1.0.0/hooks_trust.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|hooks.changed|hooks.not_configured|interaction.required|internal.contract_violation|internal.unclassified` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `confirmation_required_without_bypass`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `hooks.trust`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:hooks.trust:input`; response `urn:fbrcm:schema:cli:1.0.0:command:hooks.trust:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm hooks trust`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `hooks.trust`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/hooks/*_test.go`; E2E: 1 scenario(s): `hooks_trust_json`. |
| Verdict | **PASS.** |

### `hooks.untrust`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `hooks.untrust`; argv `fbrcm hooks untrust`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/hooks_untrust.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", ineffective); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/hooks_untrust.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `hooks command-specific result DTO`; structural schema `schemas/cli/1.0.0/hooks_untrust.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `true`; reasons `removes persisted hook trust`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `hooks.untrust`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:hooks.untrust:input`; response `urn:fbrcm:schema:cli:1.0.0:command:hooks.untrust:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm hooks untrust`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `hooks.untrust`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/hooks/*_test.go`; E2E: 1 scenario(s): `hooks_untrust_json`. |
| Verdict | **PASS.** |

### `personalizations.list`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `personalizations.list`; argv `fbrcm personalizations list`. |
| Arguments | `project` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/personalizations_list.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--update` (bool, default=false, effective_when=1 clause(s)). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/personalizations_list.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `managedfeatures command-specific DTO / core entry DTO`; structural schema `schemas/cli/1.0.0/personalizations_list.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `oauth_authorization_returns_interaction`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 2; effects `firebase_remote_read|local_cache_write|local_state_write|authentication_remote_access|local_file_write`; network `conditional` (3 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `personalizations.list`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:personalizations.list:input`; response `urn:fbrcm:schema:cli:1.0.0:command:personalizations.list:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm personalizations list`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `personalizations.list`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/managedfeatures/*_test.go`; E2E: 2 scenario(s): `personalizations_list_json`, `personalizations_list_stateless_json`. |
| Verdict | **PASS.** |

### `personalizations.show`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `personalizations.show`; argv `fbrcm personalizations show`. |
| Arguments | `project` (required; string); `personalization_id` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/personalizations_show.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--update` (bool, default=false, effective_when=1 clause(s)). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/personalizations_show.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 1 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/personalizations_show.input.schema.json`: `#/properties/arguments/properties/personalization_id/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `managedfeatures command-specific DTO / core entry DTO`; structural schema `schemas/cli/1.0.0/personalizations_show.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|personalization.not_found|profile.invalid|project.ambiguous|project.not_found|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `oauth_authorization_returns_interaction`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 2; effects `firebase_remote_read|local_cache_write|local_state_write|authentication_remote_access|local_file_write`; network `conditional` (3 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `personalizations.show`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:personalizations.show:input`; response `urn:fbrcm:schema:cli:1.0.0:command:personalizations.show:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm personalizations show`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `personalizations.show`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/managedfeatures/*_test.go`; E2E: 2 scenario(s): `personalizations_show_json`, `personalizations_show_stateless_json`. |
| Verdict | **PASS.** |

### `profile`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `profile`; argv `fbrcm profile`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/profile.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/profile.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `profile.profileCurrentResult`; structural schema `schemas/cli/1.0.0/profile.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|profile.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `profile`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:profile:input`; response `urn:fbrcm:schema:cli:1.0.0:command:profile:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm profile`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `profile`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/profile/*_test.go`; E2E: none. |
| Verdict | **PASS.** |

### `profile.delete`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `profile.delete`; argv `fbrcm profile delete`. |
| Arguments | `profile` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/profile_delete.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/profile_delete.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 1 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/profile_delete.input.schema.json`: `#/properties/arguments/properties/profile/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `profile command-specific result DTO`; structural schema `schemas/cli/1.0.0/profile_delete.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/status=deleted`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|interaction.required|internal.contract_violation|internal.unclassified|profile.conflict|profile.invalid|profile.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `confirmation_required_without_bypass`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 1; effects `local_file_delete|local_cache_delete|local_draft_delete|local_state_write`; network `none` (0 conditional clauses); destructive `true`; reasons `removes a persisted profile`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `profile.delete`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:profile.delete:input`; response `urn:fbrcm:schema:cli:1.0.0:command:profile.delete:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm profile delete`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `profile.delete`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/profile/*_test.go`; E2E: 1 scenario(s): `profile_delete_json`. |
| Verdict | **PASS.** |

### `profile.list`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `profile.list`; argv `fbrcm profile list`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/profile_list.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/profile_list.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `profile command-specific result DTO`; structural schema `schemas/cli/1.0.0/profile_list.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|profile.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `profile.list`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:profile.list:input`; response `urn:fbrcm:schema:cli:1.0.0:command:profile.list:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm profile list`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `profile.list`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/profile/*_test.go`; E2E: 1 scenario(s): `profile_list_json`. |
| Verdict | **PASS.** |

### `profile.path`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `profile.path`; argv `fbrcm profile path`. |
| Arguments | `profile` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/profile_path.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/profile_path.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `profile command-specific result DTO`; structural schema `schemas/cli/1.0.0/profile_path.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|profile.invalid|profile.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `profile.path`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:profile.path:input`; response `urn:fbrcm:schema:cli:1.0.0:command:profile.path:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm profile path`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `profile.path`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/profile/*_test.go`; E2E: 1 scenario(s): `profile_path_json`. |
| Verdict | **PASS.** |

### `profile.rename`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `profile.rename`; argv `fbrcm profile rename`. |
| Arguments | `old_name` (required; string); `new_name` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/profile_rename.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/profile_rename.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 1 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/profile_rename.input.schema.json`: `#/properties/arguments/properties/old_name/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `profile command-specific result DTO`; structural schema `schemas/cli/1.0.0/profile_rename.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|profile.conflict|profile.invalid|profile.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write|local_cache_move`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `profile.rename`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:profile.rename:input`; response `urn:fbrcm:schema:cli:1.0.0:command:profile.rename:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm profile rename`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `profile.rename`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/profile/*_test.go`; E2E: 1 scenario(s): `profile_rename_json`. |
| Verdict | **PASS.** |

### `profile.switch`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `profile.switch`; argv `fbrcm profile switch`. |
| Arguments | `name` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/profile_switch.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/profile_switch.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `profile command-specific result DTO`; structural schema `schemas/cli/1.0.0/profile_switch.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|profile.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `profile.switch`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:profile.switch:input`; response `urn:fbrcm:schema:cli:1.0.0:command:profile.switch:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm profile switch`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `profile.switch`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/profile/*_test.go`; E2E: 1 scenario(s): `profile_switch_json`. |
| Verdict | **PASS.** |

### `project.defaults`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `project.defaults`; argv `fbrcm project defaults`. |
| Arguments | `project` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/project_defaults.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--format` (string, default="json"); `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--to` (string, default=""); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/project_defaults.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 1 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/project_defaults.input.schema.json`: `#/$defs/stateless_target_selector/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `contract.ArtifactData`; structural schema `schemas/cli/1.0.0/project_defaults.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/encoding=json|utf-8|base64|none; properties/media_type=application/json|application/xml|application/x-plist`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `declared_conditions_return_interaction`; 2 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 2; effects `firebase_remote_read|local_file_write|local_state_write|authentication_remote_access`; network `required` (0 conditional clauses); destructive `true`; reasons `an existing destination file may be overwritten`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `project.defaults`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:project.defaults:input`; response `urn:fbrcm:schema:cli:1.0.0:command:project.defaults:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm project defaults`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `project.defaults`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/project/*_test.go`; E2E: 5 scenario(s): `project_defaults_file_json`, `project_defaults_json`, `project_defaults_plist`, `project_defaults_stateless_json`, `project_defaults_xml`. |
| Verdict | **PASS.** |

### `project.export`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `project.export`; argv `fbrcm project export`. |
| Arguments | `project` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/project_export.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--to` (string, default=""); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/project_export.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 1 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/project_export.input.schema.json`: `#/$defs/stateless_target_selector/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `contract.ArtifactData`; structural schema `schemas/cli/1.0.0/project_export.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/encoding=json|none`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `declared_conditions_return_interaction`; 2 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 2; effects `firebase_remote_read|local_file_write|local_state_write|authentication_remote_access`; network `required` (0 conditional clauses); destructive `true`; reasons `an existing destination file may be overwritten`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `project.export`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:project.export:input`; response `urn:fbrcm:schema:cli:1.0.0:command:project.export:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm project export`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `project.export`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/project/*_test.go`; E2E: 5 scenario(s): `project_export_file_json`, `project_export_json`, `project_export_stateless_incorrect_token_json`, `project_export_stateless_json`, `project_export_stateless_missing_token_json`. |
| Verdict | **PASS.** |

### `project.import`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `project.import`; argv `fbrcm project import`. |
| Arguments | `project` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/project_import.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--change-note` (string, default=""); `--draft` (bool, default=false, effective_when=1 clause(s)); `--dry-run` (bool, default=false); `--expr` (string, default=""); `--filter` (stringArray, default=[], aliases=-f, repeatable); `--from` (string, default=""); `--group` (stringArray, default=[], repeatable); `--keep-portable-conditions-only` (bool, default=false); `--merge` (bool, default=false); `--merge-resolve` (string, default=""); `--no-local-config` (bool, default=false); `--override` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--remove-all-conditions` (bool, default=false); `--search` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/project_import.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 4 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/project_import.input.schema.json`: `#/$defs/stateless_target_selector/x-fbrcm-matching`, `#/properties/options/properties/filter/items/x-fbrcm-matching`, `#/properties/options/properties/search/x-fbrcm-matching`, `#/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | Modes `json_document`; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:remote_config_import`; normalized property `schemas/cli/1.0.0/project_import.input.schema.json#/properties/stdin`; option restrictions are schema-enforced. |
| Success | Registered DTO type(s): `project/import.Result`; structural schema `schemas/cli/1.0.0/project_import.response.schema.json#/$defs/success_data`; reachable outcome set `success|partial_success|failure`; success variants closed variants `properties/status=unchanged|validation-failed|drafted|would-draft|imported|would-import|imported-hook-failed|imported-cache-failed`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|batch.failed|batch.partial_success|command.canceled|command.timeout|configuration.invalid|draft.exists|expression.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|group.not_found|hook.failed|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|publication.cache_failed|publication.hook_failed|remote_config.conflict|remote_config.invalid|remote_config.validation_failed|resource.not_found|stdin.remote_config.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | Codes `publication.cache_stale|publication.post_publish_hook_failed`; typed details/remediation are constrained by the shared envelope and response schema. |
| Interaction | Mode `optional`; JSON behavior `declared_conditions_return_interaction`; 5 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 3; effects `firebase_remote_read|local_cache_write|firebase_remote_write|firebase_remote_validation|local_draft_write|trusted_hook_execution|local_state_write|authentication_remote_access|local_file_write`; network `required` (0 conditional clauses); destructive `true`; reasons `selected changes may replace or remove existing values`; idempotency `conditional` (7 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `project.import`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:project.import:input`; response `urn:fbrcm:schema:cli:1.0.0:command:project.import:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`; stdin `urn:fbrcm:schema:cli:1.0.0:stdin:remote_config_import`. |
| Documentation | `docs/CLI.md § fbrcm project import`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `project.import`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/project/*_test.go`; E2E: 1 scenario(s): `project_import_stateless_json`. |
| Verdict | **PASS.** |

### `project.open`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `project.open`; argv `fbrcm project open`. |
| Arguments | `project` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/project_open.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/project_open.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `project.projectOpenResult`; structural schema `schemas/cli/1.0.0/project_open.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `browser_launch_suppressed_and_oauth_authorization_returns_interaction`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 2; effects `firebase_remote_read|local_state_write|authentication_remote_access|local_file_write`; network `conditional` (1 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `project.open`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:project.open:input`; response `urn:fbrcm:schema:cli:1.0.0:command:project.open:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm project open`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `project.open`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/project/*_test.go`; E2E: 2 scenario(s): `project_open_json`, `project_open_stateless_json`. |
| Verdict | **PASS.** |

### `project.quota-project.set`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `project.quota-project.set`; argv `fbrcm project quota-project set`. |
| Arguments | `project` (required; string); `quota_project_id` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/project_quota-project_set.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/project_quota-project_set.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `project.projectQuotaProjectResult`; structural schema `schemas/cli/1.0.0/project_quota-project_set.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/source=environment|project|auth|credentials|target; properties/status=set|unchanged`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|profile.invalid|project.ambiguous|project.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `project.quota-project.set`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.set:input`; response `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.set:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm project quota-project show|set|unset`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `project.quota-project.set`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/project/*_test.go`; E2E: 1 scenario(s): `project_quota_project_set_json`. |
| Verdict | **PASS.** |

### `project.quota-project.show`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `project.quota-project.show`; argv `fbrcm project quota-project show`. |
| Arguments | `project` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/project_quota-project_show.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/project_quota-project_show.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `project.projectQuotaProjectResult`; structural schema `schemas/cli/1.0.0/project_quota-project_show.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/source=environment|project|auth|credentials|target`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 2; effects `authentication_remote_access|local_state_write`; network `conditional` (1 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `project.quota-project.show`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.show:input`; response `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.show:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm project quota-project show|set|unset`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `project.quota-project.show`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/project/*_test.go`; E2E: 1 scenario(s): `project_quota_project_show_json`. |
| Verdict | **PASS.** |

### `project.quota-project.unset`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `project.quota-project.unset`; argv `fbrcm project quota-project unset`. |
| Arguments | `project` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/project_quota-project_unset.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/project_quota-project_unset.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `project.projectQuotaProjectResult`; structural schema `schemas/cli/1.0.0/project_quota-project_unset.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/source=environment|project|auth|credentials|target; properties/status=unset|unchanged`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 2; effects `local_state_write|authentication_remote_access`; network `conditional` (1 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `project.quota-project.unset`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.unset:input`; response `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.unset:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm project quota-project show|set|unset`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `project.quota-project.unset`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/project/*_test.go`; E2E: 1 scenario(s): `project_quota_project_unset_json`. |
| Verdict | **PASS.** |

### `project.show`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `project.show`; argv `fbrcm project show`. |
| Arguments | `project` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/project_show.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--update` (bool, default=false, effective_when=1 clause(s)). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/project_show.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `shared.ProjectJSON`; structural schema `schemas/cli/1.0.0/project_show.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/primary_template=client|server; properties/quota_project_source=environment|project|auth|credentials|target; properties/templates/items=client|server`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `oauth_authorization_returns_interaction`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 2; effects `firebase_remote_read|local_state_write|authentication_remote_access|local_file_write`; network `conditional` (3 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `project.show`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:project.show:input`; response `urn:fbrcm:schema:cli:1.0.0:command:project.show:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm project show`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `project.show`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/project/*_test.go`; E2E: 2 scenario(s): `project_show_json`, `project_show_stateless_json`. |
| Verdict | **PASS.** |

### `project.templates.set`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `project.templates.set`; argv `fbrcm project templates set`. |
| Arguments | `project` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/project_templates_set.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--primary` (string, default=""); `--profile` (string, default=""); `--stateless` (bool, default=false); `--templates` (stringSlice, default=[], repeatable); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/project_templates_set.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `project.projectTemplatesJSON`; structural schema `schemas/cli/1.0.0/project_templates_set.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/primary_template=client|server; properties/templates/items=client|server`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|profile.invalid|project.ambiguous|project.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `project.templates.set`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:project.templates.set:input`; response `urn:fbrcm:schema:cli:1.0.0:command:project.templates.set:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm project templates set`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `project.templates.set`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/project/*_test.go`; E2E: 1 scenario(s): `project_templates_set_json`. |
| Verdict | **PASS.** |

### `project.templates.show`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `project.templates.show`; argv `fbrcm project templates show`. |
| Arguments | `project` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/project_templates_show.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/project_templates_show.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `project.projectTemplatesJSON`; structural schema `schemas/cli/1.0.0/project_templates_show.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/primary_template=client|server; properties/templates/items=client|server`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|profile.invalid|project.ambiguous|project.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `project.templates.show`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:project.templates.show:input`; response `urn:fbrcm:schema:cli:1.0.0:command:project.templates.show:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm project templates show`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `project.templates.show`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/project/*_test.go`; E2E: 1 scenario(s): `project_templates_show_json`. |
| Verdict | **PASS.** |

### `projects.aliases.import`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `projects.aliases.import`; argv `fbrcm projects aliases import`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/projects_aliases_import.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--conflict` (string, default="error"); `--dry-run` (bool, default=false); `--from` (string, default="", required); `--no-local-config` (bool, default=false); `--profile` (string, default="", ineffective); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/projects_aliases_import.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `projects alias command-specific DTO`; structural schema `schemas/cli/1.0.0/projects_aliases_import.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/conflict_policy=error|keep|overwrite; properties/items/items/properties/action=add|unchanged|keep|overwrite`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|configuration.local_disabled|configuration.project_aliases_invalid|file.io_failed|filesystem.permission_denied|interaction.required|internal.contract_violation|internal.unclassified|project_alias.conflict|project_alias.read_only` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `confirmation_required_without_bypass`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `true`; reasons `--conflict overwrite may replace persisted project aliases`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `projects.aliases.import`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.import:input`; response `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.import:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm projects aliases import`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `projects.aliases.import`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/projects/*_test.go`; E2E: 1 scenario(s): `projects_aliases_import_json`. |
| Verdict | **PASS.** |

### `projects.aliases.list`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `projects.aliases.list`; argv `fbrcm projects aliases list`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/projects_aliases_list.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", ineffective); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/projects_aliases_list.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `projects alias command-specific DTO`; structural schema `schemas/cli/1.0.0/projects_aliases_list.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/items/items/properties/source=fbrcm|firebase|both`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|configuration.local_disabled|configuration.project_aliases_invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `projects.aliases.list`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.list:input`; response `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.list:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm projects aliases list`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `projects.aliases.list`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/projects/*_test.go`; E2E: 2 scenario(s): `projects_aliases_list_json`, `projects_aliases_list_narrow`. |
| Verdict | **PASS.** |

### `projects.aliases.remove`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `projects.aliases.remove`; argv `fbrcm projects aliases remove`. |
| Arguments | `alias` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/projects_aliases_remove.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", ineffective); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/projects_aliases_remove.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 1 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/projects_aliases_remove.input.schema.json`: `#/properties/arguments/properties/alias/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `projects alias command-specific DTO`; structural schema `schemas/cli/1.0.0/projects_aliases_remove.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/remaining_source=fbrcm|firebase|both; properties/source=fbrcm|firebase|both; properties/status=not_found|removed|removed_native`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|configuration.local_disabled|configuration.project_aliases_invalid|file.io_failed|filesystem.permission_denied|interaction.required|internal.contract_violation|internal.unclassified|project_alias.conflict|project_alias.read_only` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `confirmation_required_without_bypass`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `true`; reasons `removes persisted project aliases`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `projects.aliases.remove`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.remove:input`; response `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.remove:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm projects aliases remove`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `projects.aliases.remove`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/projects/*_test.go`; E2E: 1 scenario(s): `projects_aliases_remove_json`. |
| Verdict | **PASS.** |

### `projects.aliases.set`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `projects.aliases.set`; argv `fbrcm projects aliases set`. |
| Arguments | `alias` (required; string); `project_id` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/projects_aliases_set.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", ineffective); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/projects_aliases_set.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `projects alias command-specific DTO`; structural schema `schemas/cli/1.0.0/projects_aliases_set.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/source=fbrcm|firebase|both`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|configuration.local_disabled|configuration.project_aliases_invalid|file.io_failed|filesystem.permission_denied|interaction.required|internal.contract_violation|internal.unclassified|project_alias.conflict|project_alias.read_only` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `confirmation_required_without_bypass`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `true`; reasons `replaces an existing project alias when the mapping changes`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `projects.aliases.set`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.set:input`; response `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.set:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm projects aliases set`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `projects.aliases.set`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/projects/*_test.go`; E2E: 1 scenario(s): `projects_aliases_set_json`. |
| Verdict | **PASS.** |

### `projects.diff`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `projects.diff`; argv `fbrcm projects diff`. |
| Arguments | `source_project` (required; string); `target_project` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/projects_diff.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--cached` (bool, default=false, effective_when=1 clause(s)); `--conditions` (bool, default=false); `--expr` (string, default=""); `--filter` (stringArray, default=[], aliases=-f, repeatable); `--group` (stringArray, default=[], repeatable); `--no-local-config` (bool, default=false); `--parameters` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--search` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/projects_diff.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 4 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/projects_diff.input.schema.json`: `#/$defs/stateless_target_selector/x-fbrcm-matching`, `#/properties/options/properties/filter/items/x-fbrcm-matching`, `#/properties/options/properties/search/x-fbrcm-matching`, `#/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `projects command-specific result DTO / shared.PathResult`; structural schema `schemas/cli/1.0.0/projects_diff.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/changes/properties/conditions/items/properties/kind=added|removed|changed|unchanged; properties/changes/properties/group_descriptions/items/properties/kind=added|removed|changed|unchanged; properties/changes/properties/parameters/items/properties/current/properties/conditionalValues/additionalProperties/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; properties/changes/properties/parameters/items/properties/current/properties/defaultValue/oneOf/0/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; properties/changes/properties/parameters/items/properties/current/properties/valueType=|PARAMETER_VALUE_TYPE_UNSPECIFIED|STRING|BOOLEAN|NUMBER|JSON; properties/changes/properties/parameters/items/properties/final/properties/conditionalValues/additionalProperties/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; properties/changes/properties/parameters/items/properties/final/properties/defaultValue/oneOf/0/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; properties/changes/properties/parameters/items/properties/final/properties/valueType=|PARAMETER_VALUE_TYPE_UNSPECIFIED|STRING|BOOLEAN|NUMBER|JSON; properties/changes/properties/parameters/items/properties/kind=added|removed|changed|unchanged`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|expression.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|group.not_found|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `oauth_authorization_returns_interaction`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 2; effects `firebase_remote_read|local_state_write|authentication_remote_access|local_file_write`; network `conditional` (1 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `projects.diff`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:projects.diff:input`; response `urn:fbrcm:schema:cli:1.0.0:command:projects.diff:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm projects diff`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `projects.diff`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/projects/*_test.go`; E2E: 2 scenario(s): `projects_diff_client_server_json`, `projects_diff_stateless_json`. |
| Verdict | **PASS.** |

### `projects.forget`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `projects.forget`; argv `fbrcm projects forget`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/projects_forget.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--expr` (string, default=""); `--filter` (stringArray, default=[], aliases=-f, repeatable); `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/projects_forget.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 2 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/projects_forget.input.schema.json`: `#/properties/options/properties/filter/items/x-fbrcm-matching`, `#/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `projects command-specific result DTO / shared.PathResult`; structural schema `schemas/cli/1.0.0/projects_forget.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/items/items/properties/status=forgotten`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|expression.invalid|file.io_failed|filesystem.permission_denied|interaction.required|internal.contract_violation|internal.unclassified|profile.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `confirmation_required_without_bypass`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 1; effects `local_state_write|local_cache_delete|local_draft_delete`; network `none` (0 conditional clauses); destructive `true`; reasons `removes projects from the local registry and deletes their cached templates, version snapshots, and drafts`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `projects.forget`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:projects.forget:input`; response `urn:fbrcm:schema:cli:1.0.0:command:projects.forget:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm projects forget`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `projects.forget`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/projects/*_test.go`; E2E: 1 scenario(s): `projects_forget_json`. |
| Verdict | **PASS.** |

### `projects.list`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `projects.list`; argv `fbrcm projects list`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/projects_list.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--expr` (string, default=""); `--filter` (stringArray, default=[], aliases=-f, repeatable); `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--update` (bool, default=false, effective_when=1 clause(s)); `--url` (bool, default=false). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/projects_list.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 2 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/projects_list.input.schema.json`: `#/properties/options/properties/filter/items/x-fbrcm-matching`, `#/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `[]shared.ProjectJSON`; structural schema `schemas/cli/1.0.0/projects_list.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/items/items/properties/primary_template=client|server; properties/items/items/properties/quota_project_source=environment|project|auth|credentials|target; properties/items/items/properties/templates/items=client|server`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|expression.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `oauth_authorization_returns_interaction`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 2; effects `firebase_remote_read|local_state_write|authentication_remote_access|local_file_write`; network `conditional` (3 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `projects.list`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:projects.list:input`; response `urn:fbrcm:schema:cli:1.0.0:command:projects.list:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm projects list`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `projects.list`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/projects/*_test.go`; E2E: 3 scenario(s): `projects_list_json`, `projects_list_stateless_expr_json`, `projects_list_stateless_json`. |
| Verdict | **PASS.** |

### `projects.path`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `projects.path`; argv `fbrcm projects path`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/projects_path.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/projects_path.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `projects command-specific result DTO / shared.PathResult`; structural schema `schemas/cli/1.0.0/projects_path.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|profile.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `projects.path`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:projects.path:input`; response `urn:fbrcm:schema:cli:1.0.0:command:projects.path:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm projects path`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `projects.path`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/projects/*_test.go`; E2E: 1 scenario(s): `projects_path_json`. |
| Verdict | **PASS.** |

### `projects.promote`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `projects.promote`; argv `fbrcm projects promote`. |
| Arguments | `source_project` (required; string); `target_project` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/projects_promote.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--all` (bool, default=false); `--change-note` (string, default=""); `--conditions` (bool, default=false); `--dry-run` (bool, default=false); `--expr` (string, default=""); `--filter` (stringArray, default=[], aliases=-f, repeatable); `--group` (stringArray, default=[], repeatable); `--interactive` (bool, default=false); `--no-local-config` (bool, default=false); `--parameters` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--prune` (bool, default=false); `--search` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/projects_promote.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 4 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/projects_promote.input.schema.json`: `#/$defs/stateless_target_selector/x-fbrcm-matching`, `#/properties/options/properties/filter/items/x-fbrcm-matching`, `#/properties/options/properties/search/x-fbrcm-matching`, `#/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `projects command-specific result DTO / shared.PathResult`; structural schema `schemas/cli/1.0.0/projects_promote.response.schema.json#/$defs/success_data`; reachable outcome set `success|partial_success|failure`; success variants closed variants `properties/status=unchanged|would-publish|published|failed|validation-failed|published-hook-failed|published-cache-failed`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|batch.failed|batch.partial_success|command.canceled|command.timeout|configuration.invalid|draft.exists|expression.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|group.not_found|hook.failed|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|publication.cache_failed|publication.hook_failed|remote_config.conflict|remote_config.invalid|remote_config.validation_failed|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | Codes `publication.cache_stale|publication.post_publish_hook_failed`; typed details/remediation are constrained by the shared envelope and response schema. |
| Interaction | Mode `optional`; JSON behavior `declared_conditions_return_interaction`; 3 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 3; effects `firebase_remote_read|local_cache_write|firebase_remote_write|firebase_remote_validation|trusted_hook_execution|local_state_write|authentication_remote_access|local_file_write`; network `required` (0 conditional clauses); destructive `true`; reasons `selected changes may replace or remove existing values`; idempotency `conditional` (6 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `projects.promote`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:projects.promote:input`; response `urn:fbrcm:schema:cli:1.0.0:command:projects.promote:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm projects promote`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `projects.promote`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/projects/*_test.go`; E2E: 1 scenario(s): `projects_promote_stateless_dry_run_json`. |
| Verdict | **PASS.** |

### `projects.reset`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `projects.reset`; argv `fbrcm projects reset`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/projects_reset.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/projects_reset.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `projects command-specific result DTO / shared.PathResult`; structural schema `schemas/cli/1.0.0/projects_reset.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/status=reset`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|interaction.required|internal.contract_violation|internal.unclassified|profile.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `confirmation_required_without_bypass`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 1; effects `local_state_write|local_file_delete`; network `none` (0 conditional clauses); destructive `true`; reasons `replaces the local project registry`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `projects.reset`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:projects.reset:input`; response `urn:fbrcm:schema:cli:1.0.0:command:projects.reset:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm projects reset`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `projects.reset`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/projects/*_test.go`; E2E: 1 scenario(s): `projects_reset_json`. |
| Verdict | **PASS.** |

### `projects.update`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `projects.update`; argv `fbrcm projects update`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/projects_update.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--auth` (string, default=""); `--expr` (string, default=""); `--filter` (stringArray, default=[], aliases=-f, repeatable); `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--url` (bool, default=false). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/projects_update.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 2 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/projects_update.input.schema.json`: `#/properties/options/properties/filter/items/x-fbrcm-matching`, `#/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `[]shared.ProjectJSON`; structural schema `schemas/cli/1.0.0/projects_update.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/items/items/properties/primary_template=client|server; properties/items/items/properties/quota_project_source=environment|project|auth|credentials|target; properties/items/items/properties/templates/items=client|server`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|expression.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `oauth_authorization_returns_interaction`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 2; effects `firebase_remote_read|local_state_write|authentication_remote_access|local_file_write`; network `required` (0 conditional clauses); destructive `false`; idempotency `no` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `projects.update`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:projects.update:input`; response `urn:fbrcm:schema:cli:1.0.0:command:projects.update:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm projects update`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `projects.update`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/projects/*_test.go`; E2E: 1 scenario(s): `projects_update_json`. |
| Verdict | **PASS.** |

### `rollouts.delete`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `rollouts.delete`; argv `fbrcm rollouts delete`. |
| Arguments | `project` (required; string); `rollout_id` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/rollouts_delete.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/rollouts_delete.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `managedfeatures command-specific DTO / core entry DTO`; structural schema `schemas/cli/1.0.0/rollouts_delete.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/status=deleted|would-delete`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `declared_conditions_return_interaction`; 2 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 3; effects `firebase_remote_read|firebase_managed_feature_delete|local_state_write|authentication_remote_access|local_file_write`; network `required` (0 conditional clauses); destructive `true`; reasons `removes a Firebase Remote Config rollout`; idempotency `conditional` (3 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `rollouts.delete`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:rollouts.delete:input`; response `urn:fbrcm:schema:cli:1.0.0:command:rollouts.delete:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm rollouts delete`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `rollouts.delete`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/managedfeatures/*_test.go`; E2E: 1 scenario(s): `rollouts_delete_stateless_confirmation_json`. |
| Verdict | **PASS.** |

### `rollouts.list`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `rollouts.list`; argv `fbrcm rollouts list`. |
| Arguments | `project` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/rollouts_list.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--update` (bool, default=false, effective_when=1 clause(s)). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/rollouts_list.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `managedfeatures command-specific DTO / core entry DTO`; structural schema `schemas/cli/1.0.0/rollouts_list.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `oauth_authorization_returns_interaction`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 2; effects `firebase_remote_read|local_cache_write|local_state_write|authentication_remote_access|local_file_write`; network `required` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `rollouts.list`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:rollouts.list:input`; response `urn:fbrcm:schema:cli:1.0.0:command:rollouts.list:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm rollouts list`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `rollouts.list`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/managedfeatures/*_test.go`; E2E: 2 scenario(s): `rollouts_list_json`, `rollouts_list_stateless_json`. |
| Verdict | **PASS.** |

### `rollouts.show`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `rollouts.show`; argv `fbrcm rollouts show`. |
| Arguments | `project` (required; string); `rollout_id` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/rollouts_show.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--update` (bool, default=false, effective_when=1 clause(s)). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/rollouts_show.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `managedfeatures command-specific DTO / core entry DTO`; structural schema `schemas/cli/1.0.0/rollouts_show.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|resource.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `oauth_authorization_returns_interaction`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 2; effects `firebase_remote_read|local_cache_write|local_state_write|authentication_remote_access|local_file_write`; network `required` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `rollouts.show`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:rollouts.show:input`; response `urn:fbrcm:schema:cli:1.0.0:command:rollouts.show:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm rollouts show`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `rollouts.show`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/managedfeatures/*_test.go`; E2E: 2 scenario(s): `rollouts_show_json`, `rollouts_show_stateless_json`. |
| Verdict | **PASS.** |

### `root`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `root`; argv `fbrcm`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/root.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false, effective_when=1 clause(s)); `--profile` (string, default="", effective_when=1 clause(s)); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--version` (bool, default=false, aliases=-v). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/root.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `contract.CapabilityIndex | contract.TextData`; structural schema `schemas/cli/1.0.0/root.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|argument.unknown_command|command.canceled|command.not_executable|command.not_found|command.timeout|configuration.invalid|internal.contract_violation|internal.unclassified|profile.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `root`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:root:input`; response `urn:fbrcm:schema:cli:1.0.0:command:root:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `root`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/app/*_test.go`; E2E: 2 scenario(s): `root_version`, `root_version_json`. |
| Verdict | **PASS.** |

### `schema.list`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `schema.list`; argv `fbrcm schema list`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/schema_list.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", ineffective); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/schema_list.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `meta.schemaListResult`; structural schema `schemas/cli/1.0.0/schema_list.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|internal.contract_violation` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `schema.list`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:schema.list:input`; response `urn:fbrcm:schema:cli:1.0.0:command:schema.list:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm schema list`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `schema.list`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/app/*_test.go`; E2E: 1 scenario(s): `schema_list_json`. |
| Verdict | **PASS.** |

### `schema.show`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `schema.show`; argv `fbrcm schema show`. |
| Arguments | `schema_id` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/schema_show.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", ineffective); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/schema_show.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 1 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/schema_show.input.schema.json`: `#/properties/arguments/properties/schema_id/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `contract.JSONDocument`; structural schema `schemas/cli/1.0.0/schema_show.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|internal.contract_violation|schema.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `schema.show`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:schema.show:input`; response `urn:fbrcm:schema:cli:1.0.0:command:schema.show:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm schema show`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `schema.show`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/app/*_test.go`; E2E: 1 scenario(s): `schema_show_json`. |
| Verdict | **PASS.** |

### `theme`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `theme`; argv `fbrcm theme`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/theme.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", ineffective); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/theme.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `theme.themeCurrentResult`; structural schema `schemas/cli/1.0.0/theme.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/source=default|global|local`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|profile.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `theme`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:theme:input`; response `urn:fbrcm:schema:cli:1.0.0:command:theme:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm theme`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `theme`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/theme/*_test.go`; E2E: 1 scenario(s): `theme_current_json`. |
| Verdict | **PASS.** |

### `theme.delete`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `theme.delete`; argv `fbrcm theme delete`. |
| Arguments | `theme` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/theme_delete.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", ineffective); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/theme_delete.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 1 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/theme_delete.input.schema.json`: `#/properties/arguments/properties/theme/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `theme command-specific result DTO`; structural schema `schemas/cli/1.0.0/theme_delete.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/status=deleted`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|interaction.required|internal.contract_violation|internal.unclassified|theme.conflict|theme.invalid|theme.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `confirmation_required_without_bypass`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 1; effects `local_file_delete|local_state_write`; network `none` (0 conditional clauses); destructive `true`; reasons `removes an installed theme file`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `theme.delete`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:theme.delete:input`; response `urn:fbrcm:schema:cli:1.0.0:command:theme.delete:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm theme delete`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `theme.delete`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/theme/*_test.go`; E2E: none. |
| Verdict | **PASS.** |

### `theme.import`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `theme.import`; argv `fbrcm theme import`. |
| Arguments | `source` (optional; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/theme_import.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--name` (string, default=""); `--no-local-config` (bool, default=false); `--profile` (string, default="", ineffective); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/theme_import.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | Modes `toml_document`; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:theme`; normalized property `schemas/cli/1.0.0/theme_import.input.schema.json#/properties/stdin`; option restrictions are schema-enforced. |
| Success | Registered DTO type(s): `theme command-specific result DTO`; structural schema `schemas/cli/1.0.0/theme_import.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `oneOf/0/properties/status=imported; oneOf/1/properties/items/items/properties/reason=already_exists; oneOf/1/properties/items/items/properties/status=imported|skipped`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|stdin.remote_config.invalid|theme.conflict|theme.invalid|validation.failed` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | Codes `theme.already_exists`; typed details/remediation are constrained by the shared envelope and response schema. |
| Interaction | Mode `optional`; JSON behavior `missing_input_returns_interaction`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 2; effects `local_file_write|local_state_write`; network `conditional` (1 conditional clauses); destructive `false`; idempotency `conditional` (2 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `theme.import`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:theme.import:input`; response `urn:fbrcm:schema:cli:1.0.0:command:theme.import:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`; stdin `urn:fbrcm:schema:cli:1.0.0:stdin:theme`. |
| Documentation | `docs/CLI.md § fbrcm theme import`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `theme.import`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/theme/*_test.go`; E2E: 2 scenario(s): `theme_import_directory_json`, `theme_import_json`. |
| Verdict | **PASS.** |

### `theme.list`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `theme.list`; argv `fbrcm theme list`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/theme_list.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", ineffective); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/theme_list.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `theme command-specific result DTO`; structural schema `schemas/cli/1.0.0/theme_list.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `theme.list`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:theme.list:input`; response `urn:fbrcm:schema:cli:1.0.0:command:theme.list:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm theme list`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `theme.list`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/theme/*_test.go`; E2E: 1 scenario(s): `theme_list_json`. |
| Verdict | **PASS.** |

### `theme.path`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `theme.path`; argv `fbrcm theme path`. |
| Arguments | `theme` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/theme_path.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", ineffective); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/theme_path.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `theme command-specific result DTO`; structural schema `schemas/cli/1.0.0/theme_path.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `theme.path`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:theme.path:input`; response `urn:fbrcm:schema:cli:1.0.0:command:theme.path:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm theme path`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `theme.path`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/theme/*_test.go`; E2E: 1 scenario(s): `theme_path_json`. |
| Verdict | **PASS.** |

### `theme.rename`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `theme.rename`; argv `fbrcm theme rename`. |
| Arguments | `old_name` (required; string); `new_name` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/theme_rename.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", ineffective); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/theme_rename.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 1 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/theme_rename.input.schema.json`: `#/properties/arguments/properties/old_name/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `theme command-specific result DTO`; structural schema `schemas/cli/1.0.0/theme_rename.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|theme.conflict|theme.invalid|theme.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write|local_file_move`; network `none` (0 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `theme.rename`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:theme.rename:input`; response `urn:fbrcm:schema:cli:1.0.0:command:theme.rename:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm theme rename`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `theme.rename`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/theme/*_test.go`; E2E: 1 scenario(s): `theme_rename_json`. |
| Verdict | **PASS.** |

### `theme.reset`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `theme.reset`; argv `fbrcm theme reset`. |
| Arguments | No positional arguments (`additionalProperties:false`; exact zero arity). Arity and normalized presence are enforced by `schemas/cli/1.0.0/theme_reset.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", ineffective); `--scope` (string, default="global"); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/theme_reset.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | N/A; no selector matching annotations are published. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `theme command-specific result DTO`; structural schema `schemas/cli/1.0.0/theme_reset.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/scope=global|local; properties/status=reset|unchanged`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `true`; reasons `removes a persisted theme selection`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `theme.reset`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:theme.reset:input`; response `urn:fbrcm:schema:cli:1.0.0:command:theme.reset:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm theme reset`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `theme.reset`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/theme/*_test.go`; E2E: none. |
| Verdict | **PASS.** |

### `theme.switch`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `theme.switch`; argv `fbrcm theme switch`. |
| Arguments | `name` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/theme_switch.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--no-local-config` (bool, default=false); `--profile` (string, default="", ineffective); `--scope` (string, default="global"); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/theme_switch.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 1 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/theme_switch.input.schema.json`: `#/properties/arguments/properties/name/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `theme command-specific result DTO`; structural schema `schemas/cli/1.0.0/theme_switch.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/scope=global|local`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|internal.contract_violation|internal.unclassified|theme.invalid|theme.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | N/A; mode `none`; JSON mode has no prompt/launch branch. |
| Effects | Level 1; effects `local_state_write`; network `none` (0 conditional clauses); destructive `true`; reasons `the built-in selector removes a persisted theme selection`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `theme.switch`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:theme.switch:input`; response `urn:fbrcm:schema:cli:1.0.0:command:theme.switch:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm theme switch`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `theme.switch`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/theme/*_test.go`; E2E: none. |
| Verdict | **PASS.** |

### `update`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `update`; argv `fbrcm update`. |
| Arguments | `parameter` (optional; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/update.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--change-note` (string, default=""); `--condition` (string, default=""); `--description` (string, default=""); `--draft` (bool, default=false, effective_when=1 clause(s)); `--dry-run` (bool, default=false); `--expr` (string, default=""); `--filter` (stringArray, default=[], aliases=-f, repeatable); `--group` (string, default=""); `--name` (string, default=""); `--no-group` (bool, default=false); `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--project` (stringArray, default=[], aliases=-p, repeatable); `--remove-all-conditional-values` (bool, default=false); `--remove-conditional-value` (stringArray, default=[], repeatable); `--search` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--type` (string, default=""); `--use-in-app-default` (bool, default=false); `--value` (string, default=""); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/update.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 7 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/update.input.schema.json`: `#/$defs/stateless_get_project_selector/oneOf/0/x-fbrcm-matching`, `#/$defs/stateless_get_project_selector/oneOf/1/x-fbrcm-matching`, `#/properties/arguments/properties/parameter/x-fbrcm-matching`, `#/properties/options/properties/condition/x-fbrcm-matching`, `#/properties/options/properties/filter/items/x-fbrcm-matching`, `#/properties/options/properties/search/x-fbrcm-matching`, `#/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | Modes `json_document`; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:remote_config`; normalized property `schemas/cli/1.0.0/update.input.schema.json#/properties/stdin`; option restrictions are schema-enforced. |
| Success | Registered DTO type(s): `[]shared/rc.RemoteMutationJSONResult | contract.ArtifactData`; structural schema `schemas/cli/1.0.0/update.response.schema.json#/$defs/success_data`; reachable outcome set `success|partial_success|failure`; success variants closed variants `oneOf/0/properties/items/items/allOf/9/then/properties/error/properties/stage=publication|pre_publish_hook; oneOf/0/properties/items/items/allOf/11/then/properties/error/properties/stage=preparation|validation|publication; oneOf/1/$defs/artifact_json_content/properties/parameterGroups/additionalProperties/properties/parameters/additionalProperties/properties/conditionalValues/additionalProperties/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; oneOf/1/$defs/artifact_json_content/properties/parameterGroups/additionalProperties/properties/parameters/additionalProperties/properties/defaultValue/oneOf/0/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; oneOf/1/$defs/artifact_json_content/properties/parameterGroups/additionalProperties/properties/parameters/additionalProperties/properties/valueType=|PARAMETER_VALUE_TYPE_UNSPECIFIED|STRING|BOOLEAN|NUMBER|JSON; oneOf/1/$defs/artifact_json_content/properties/parameters/additionalProperties/properties/conditionalValues/additionalProperties/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; oneOf/1/$defs/artifact_json_content/properties/parameters/additionalProperties/properties/defaultValue/oneOf/0/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; oneOf/1/$defs/artifact_json_content/properties/parameters/additionalProperties/properties/valueType=|PARAMETER_VALUE_TYPE_UNSPECIFIED|STRING|BOOLEAN|NUMBER|JSON; oneOf/1/properties/encoding=json`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|batch.failed|batch.partial_success|command.canceled|command.timeout|configuration.invalid|draft.exists|expression.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|hook.failed|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|parameter.ambiguous|parameter.not_found|profile.invalid|project.ambiguous|project.not_found|publication.cache_failed|publication.hook_failed|remote_config.conflict|remote_config.invalid|remote_config.validation_failed|resource.not_found|stdin.remote_config.invalid` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | Codes `publication.non_atomic|publication.cache_stale|publication.post_publish_hook_failed`; typed details/remediation are constrained by the shared envelope and response schema. |
| Interaction | Mode `optional`; JSON behavior `declared_conditions_return_interaction`; 2 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 3; effects `firebase_remote_read|local_cache_write|firebase_remote_write|firebase_remote_validation|local_draft_write|trusted_hook_execution|local_state_write|authentication_remote_access|local_file_write`; network `conditional` (1 conditional clauses); destructive `true`; reasons `specific flags or selected changes may replace or remove existing values`; idempotency `conditional` (8 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `update`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:update:input`; response `urn:fbrcm:schema:cli:1.0.0:command:update:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`; stdin `urn:fbrcm:schema:cli:1.0.0:stdin:remote_config`. |
| Documentation | `docs/CLI.md § fbrcm update`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `update`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/update/*_test.go`; E2E: 3 scenario(s): `parameter_update_dry_run_json`, `parameter_update_json`, `parameter_update_stateless_json`. |
| Verdict | **PASS.** |

### `versions.diff`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `versions.diff`; argv `fbrcm versions diff`. |
| Arguments | `project` (required; string); `from` (required; string); `to` (optional; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/versions_diff.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--cached` (bool, default=false, effective_when=1 clause(s)); `--conditions` (bool, default=false); `--expr` (string, default=""); `--filter` (stringArray, default=[], aliases=-f, repeatable); `--group` (stringArray, default=[], repeatable); `--no-local-config` (bool, default=false); `--parameters` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--search` (string, default=""); `--side-by-side` (bool, default=false); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/versions_diff.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 5 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/versions_diff.input.schema.json`: `#/$defs/stateless_target_selector/x-fbrcm-matching`, `#/$defs/version_selector/x-fbrcm-matching`, `#/properties/options/properties/filter/items/x-fbrcm-matching`, `#/properties/options/properties/search/x-fbrcm-matching`, `#/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `versions.versionDiffResult`; structural schema `schemas/cli/1.0.0/versions_diff.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `properties/diff/properties/conditions/items/properties/kind=added|removed|changed|unchanged; properties/diff/properties/group_descriptions/items/properties/kind=added|removed|changed|unchanged; properties/diff/properties/parameters/items/properties/current/properties/conditionalValues/additionalProperties/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; properties/diff/properties/parameters/items/properties/current/properties/defaultValue/oneOf/0/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; properties/diff/properties/parameters/items/properties/current/properties/valueType=|PARAMETER_VALUE_TYPE_UNSPECIFIED|STRING|BOOLEAN|NUMBER|JSON; properties/diff/properties/parameters/items/properties/final/properties/conditionalValues/additionalProperties/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; properties/diff/properties/parameters/items/properties/final/properties/defaultValue/oneOf/0/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; properties/diff/properties/parameters/items/properties/final/properties/valueType=|PARAMETER_VALUE_TYPE_UNSPECIFIED|STRING|BOOLEAN|NUMBER|JSON; properties/diff/properties/parameters/items/properties/kind=added|removed|changed|unchanged`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|expression.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|resource.not_found|version.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `oauth_authorization_returns_interaction`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 2; effects `firebase_remote_read|local_cache_write|local_state_write|authentication_remote_access|local_file_write`; network `conditional` (2 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `versions.diff`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:versions.diff:input`; response `urn:fbrcm:schema:cli:1.0.0:command:versions.diff:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm versions diff`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `versions.diff`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/versions/*_test.go`; E2E: 3 scenario(s): `versions_diff_filtered_json`, `versions_diff_json`, `versions_diff_stateless_json`. |
| Verdict | **PASS.** |

### `versions.export`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `versions.export`; argv `fbrcm versions export`. |
| Arguments | `project` (required; string); `version` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/versions_export.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--cached` (bool, default=false, effective_when=1 clause(s)); `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--to` (string, default=""); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/versions_export.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 2 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/versions_export.input.schema.json`: `#/$defs/stateless_target_selector/x-fbrcm-matching`, `#/$defs/version_selector/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `contract.ArtifactData`; structural schema `schemas/cli/1.0.0/versions_export.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants closed variants `$defs/artifact_json_content/properties/parameterGroups/additionalProperties/properties/parameters/additionalProperties/properties/conditionalValues/additionalProperties/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; $defs/artifact_json_content/properties/parameterGroups/additionalProperties/properties/parameters/additionalProperties/properties/defaultValue/oneOf/0/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; $defs/artifact_json_content/properties/parameterGroups/additionalProperties/properties/parameters/additionalProperties/properties/valueType=|PARAMETER_VALUE_TYPE_UNSPECIFIED|STRING|BOOLEAN|NUMBER|JSON; $defs/artifact_json_content/properties/parameters/additionalProperties/properties/conditionalValues/additionalProperties/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; $defs/artifact_json_content/properties/parameters/additionalProperties/properties/defaultValue/oneOf/0/oneOf/5/propertyNames/not=value|useInAppDefault|personalizationValue|experimentValue|rolloutValue; $defs/artifact_json_content/properties/parameters/additionalProperties/properties/valueType=|PARAMETER_VALUE_TYPE_UNSPECIFIED|STRING|BOOLEAN|NUMBER|JSON; properties/encoding=json|none`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|resource.not_found|version.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `declared_conditions_return_interaction`; 2 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 2; effects `firebase_remote_read|local_cache_write|local_file_write|local_state_write|authentication_remote_access`; network `conditional` (2 conditional clauses); destructive `true`; reasons `an existing destination file may be overwritten`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `versions.export`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:versions.export:input`; response `urn:fbrcm:schema:cli:1.0.0:command:versions.export:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm versions export`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `versions.export`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/versions/*_test.go`; E2E: 2 scenario(s): `versions_export_file_json`, `versions_export_stateless_file_json`. |
| Verdict | **PASS.** |

### `versions.list`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `versions.list`; argv `fbrcm versions list`. |
| Arguments | `project` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/versions_list.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--all` (bool, default=false); `--before` (string, default=""); `--cached` (bool, default=false, effective_when=1 clause(s)); `--limit` (int, default=20); `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--since` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--until` (string, default=""). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/versions_list.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 1 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/versions_list.input.schema.json`: `#/$defs/stateless_target_selector/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `[]versions.versionJSON`; structural schema `schemas/cli/1.0.0/versions_list.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|resource.not_found|version.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `oauth_authorization_returns_interaction`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 2; effects `firebase_remote_read|local_state_write|authentication_remote_access|local_file_write`; network `conditional` (1 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `versions.list`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:versions.list:input`; response `urn:fbrcm:schema:cli:1.0.0:command:versions.list:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm versions list`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `versions.list`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/versions/*_test.go`; E2E: 2 scenario(s): `versions_list`, `versions_list_stateless_json`. |
| Verdict | **PASS.** |

### `versions.restore`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `versions.restore`; argv `fbrcm versions restore`. |
| Arguments | `project` (required; string); `version` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/versions_restore.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--change-note` (string, default=""); `--dry-run` (bool, default=false); `--no-local-config` (bool, default=false); `--profile` (string, default=""); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/versions_restore.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 1 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/versions_restore.input.schema.json`: `#/$defs/version_selector/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `versions.versionPublishResult`; structural schema `schemas/cli/1.0.0/versions_restore.response.schema.json#/$defs/success_data`; reachable outcome set `success|partial_success|failure`; success variants closed variants `properties/status=unchanged|would-publish|published|failed|validation-failed|published-local-update-failed|published-hook-failed|published-cache-failed`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|batch.failed|batch.partial_success|command.canceled|command.timeout|configuration.invalid|draft.exists|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|hook.failed|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|publication.cache_failed|publication.hook_failed|remote_config.conflict|remote_config.invalid|remote_config.validation_failed|resource.not_found|version.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | Codes `publication.cache_stale|publication.post_publish_hook_failed`; typed details/remediation are constrained by the shared envelope and response schema. |
| Interaction | Mode `optional`; JSON behavior `declared_conditions_return_interaction`; 2 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 3; effects `firebase_remote_read|local_cache_write|firebase_remote_write|firebase_remote_validation|trusted_hook_execution|local_state_write|authentication_remote_access|local_file_write`; network `required` (0 conditional clauses); destructive `true`; reasons `replaces the current Remote Config template`; idempotency `conditional` (6 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `versions.restore`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:versions.restore:input`; response `urn:fbrcm:schema:cli:1.0.0:command:versions.restore:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm versions restore`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `versions.restore`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/versions/*_test.go`; E2E: none. |
| Verdict | **PASS.** |

### `versions.rollback`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `versions.rollback`; argv `fbrcm versions rollback`. |
| Arguments | `project` (required; string); `version` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/versions_rollback.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--dry-run` (bool, default=false); `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"); `--yes` (bool, default=false, aliases=-y). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/versions_rollback.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 2 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/versions_rollback.input.schema.json`: `#/$defs/stateless_target_selector/x-fbrcm-matching`, `#/$defs/version_selector/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `versions.versionPublishResult`; structural schema `schemas/cli/1.0.0/versions_rollback.response.schema.json#/$defs/success_data`; reachable outcome set `success|partial_success|failure`; success variants closed variants `properties/status=unchanged|would-publish|published|failed|validation-failed|published-local-update-failed|published-hook-failed|published-cache-failed`. |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|batch.failed|batch.partial_success|command.canceled|command.timeout|configuration.invalid|draft.exists|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|hook.failed|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|publication.cache_failed|publication.hook_failed|remote_config.conflict|remote_config.invalid|remote_config.validation_failed|resource.not_found|version.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | Codes `publication.cache_stale|publication.post_publish_hook_failed`; typed details/remediation are constrained by the shared envelope and response schema. |
| Interaction | Mode `optional`; JSON behavior `declared_conditions_return_interaction`; 2 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 3; effects `firebase_remote_read|local_cache_write|firebase_remote_write|firebase_remote_validation|trusted_hook_execution|local_state_write|authentication_remote_access|local_file_write`; network `required` (0 conditional clauses); destructive `true`; reasons `replaces the current Remote Config template`; idempotency `conditional` (6 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `versions.rollback`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:versions.rollback:input`; response `urn:fbrcm:schema:cli:1.0.0:command:versions.rollback:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm versions rollback`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `versions.rollback`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/versions/*_test.go`; E2E: 1 scenario(s): `versions_rollback_stateless_dry_run_json`. |
| Verdict | **PASS.** |

### `versions.show`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `versions.show`; argv `fbrcm versions show`. |
| Arguments | `project` (required; string); `version` (required; string). Arity and normalized presence are enforced by `schemas/cli/1.0.0/versions_show.input.schema.json#/properties/arguments`; argument grammars/normalization are at the individual properties and referenced semantic definitions. |
| Options | `--cached` (bool, default=false, effective_when=1 clause(s)); `--no-local-config` (bool, default=false); `--profile` (string, default="", effective_when=1 clause(s)); `--stateless` (bool, default=false); `--timeout` (duration, default="0s"). Types, defaults, aliases, repeatability, required state, effectiveness, and all dependencies/exclusions are fixed by `schemas/cli/1.0.0/versions_show.input.schema.json#/properties/options` plus its `allOf` constraints. |
| Selection | 2 machine-readable `x-fbrcm-matching` rule location(s) in `schemas/cli/1.0.0/versions_show.input.schema.json`: `#/$defs/stateless_target_selector/x-fbrcm-matching`, `#/$defs/version_selector/x-fbrcm-matching`; these publish grammar, normalization, exactness/case, precedence, repeated/cross-source composition, omitted defaults, prefixes/canonicalization, and typed lookup boundaries. |
| Stdin | N/A; `supports.stdin=false`, no modes/schema, and normalized `stdin` is null-only. |
| Success | Registered DTO type(s): `versions.versionShowResult`; structural schema `schemas/cli/1.0.0/versions_show.response.schema.json#/$defs/success_data`; reachable outcome set `success|failure`; success variants the schema's required/nullable/omitted field variants (no closed DTO enum). |
| Failure | Typed `Problem` with code/category/details/retryability/target/stage/remediation and semantic exit status via `urn:fbrcm:schema:cli:1.0.0:error`; command-reachable top-level codes are closed to `argument.invalid|auth.configuration_invalid|auth.credentials_invalid|auth.not_found|auth.quota_project_required|auth.setup_required|command.canceled|command.timeout|configuration.invalid|file.io_failed|filesystem.permission_denied|firebase.permission_denied|firebase.rate_limited|firebase.request_failed|firebase.service_unavailable|firebase.timeout|interaction.required|internal.contract_violation|internal.unclassified|network.offline|network.timeout|network.unavailable|profile.invalid|project.ambiguous|project.not_found|resource.not_found|version.not_found` by the detailed capability and response schema. Nested batch-target failures are the documented open extension point. |
| Warnings | N/A; response schema requires an empty warning array. |
| Interaction | Mode `optional`; JSON behavior `oauth_authorization_returns_interaction`; 1 trigger clause(s) in the detailed capability return typed interaction instead of prompting/launching; bypass/required options are named in those predicates and problem details. |
| Effects | Level 2; effects `firebase_remote_read|local_cache_write|local_state_write|authentication_remote_access|local_file_write`; network `conditional` (2 conditional clauses); destructive `false`; idempotency `yes` (0 conditional clauses). Exact per-effect predicates: `cli/app/testdata/contract_v1_capabilities_detailed.golden.json` record `versions.show`. |
| Schemas | Input `urn:fbrcm:schema:cli:1.0.0:command:versions.show:input`; response `urn:fbrcm:schema:cli:1.0.0:command:versions.show:response`; shared envelope `urn:fbrcm:schema:cli:1.0.0:envelope`; error `urn:fbrcm:schema:cli:1.0.0:error`. |
| Documentation | `docs/CLI.md § fbrcm versions show`; shared JSON rules in `docs/CLI.md:367`, `docs/CLI.md:787`, and `docs/cli-contract.md`. |
| Evidence | Exact per-class entries: `cli/app/testdata/contract_v1_audit_evidence.golden.json` record `versions.show`; matrix/reference/applicability enforcement: `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix`; command tests: `cli/commands/versions/*_test.go`; E2E: 2 scenario(s): `versions_show_json`, `versions_show_stateless_json`. |
| Verdict | **PASS.** |

## Findings

None. All four findings from the 2026-08-27 audit are closed:

- F-001 (`OUT-04`): each quota-project command now publishes only its reachable status values, enforced by `TestQuotaProjectResponseSchemasConstrainCommandReachableStatuses`.
- F-002 (`ERR-03`): detailed capabilities publish command-specific `problem_codes`, response schemas close top-level errors to the same set, and `TestCommandResponseSchemasConstrainReachableProblemCodes` rejects disagreement and inapplicable sibling codes.
- F-003 (`GEN-03`): the maintained `docs/CLI.md` command inventory is compared exactly with executable set `E` by `TestEveryExecutableCommandHasDocumentationInventoryEntry`.
- F-004 (`GEN-04`): `contract_v1_audit_evidence.golden.json` contains all 108 × 15 command/test-class cells, and `TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix` rejects missing cells, unjustified applicability, stale tests, and stale E2E scenario references.

## Determinism and repository checks

| Check | Result |
| --- | --- |
| `go run ./cmd/schemagen` twice | PASS; baseline, first run, and second run SHA-256 manifests were byte-identical; generated diff empty. |
| Compile every schema `$id` | PASS; 226 Draft 2020-12 schemas, unique IDs, all references resolved. |
| `go test -count=1 ./...` (root module) | PASS. |
| `go test -count=1 ./...` (`e2e` module) | PASS with unrestricted loopback; 159 replay scenarios plus harness tests. |
| `go vet ./...` (root and `e2e`) | PASS. |
| `golangci-lint run` (root and `e2e`) | PASS; zero issues. |
| Final diff | Only intended runtime, contract, generated-schema/golden, documentation, E2E snapshot, and audit-report changes. |

All twelve acceptance conditions pass, so the required verdict is **AUTHORITATIVE: PASS**.
