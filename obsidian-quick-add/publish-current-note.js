module.exports = {
  entry: publishCurrentNote,
  settings: {
    name: "Post Sync Publish Current Note",
    author: "Codex",
    options: {
      apiBaseUrl: {
        type: "text",
        defaultValue: "http://localhost:8080/api/v1",
        placeholder: "http://localhost:8080/api/v1",
        name: "API Base URL",
        description: "post-sync backend API base URL.",
      },
      aliasMappingPath: {
        type: "text",
        defaultValue: "obsidian-quick-add/target-aliases.json",
        placeholder: "obsidian-quick-add/target-aliases.json",
        name: "Alias Mapping Path",
        description: "Vault-relative JSON file mapping alias to post-sync target IDs.",
      },
      templateFallback: {
        type: "text",
        defaultValue: "default",
        placeholder: "default",
        name: "Default Template",
        description: "Fallback template name when post_template is not set.",
      },
      showNotice: {
        type: "toggle",
        defaultValue: true,
        name: "Show Notice",
        description: "Show Obsidian notices for success or failure.",
      },
    },
  },
};

async function publishCurrentNote(params, settings) {
  const { app, obsidian } = params;
  const activeFile = app.workspace.getActiveFile();
  if (!activeFile) {
    throw new Error("No active file found");
  }
  if (activeFile.extension !== "md") {
    throw new Error("Active file must be a Markdown file");
  }

  const rawMarkdown = await app.vault.read(activeFile);
  const frontmatter = readFrontmatter(app, activeFile);
  validatePostConfig(frontmatter);

  const aliasMap = await readAliasMap(app, settings.aliasMappingPath);
  const targetIds = resolveTargetIds(frontmatter.post_targets, aliasMap);
  const uploadMarkdown = buildUploadMarkdown(frontmatter, rawMarkdown, obsidian);

  const apiBaseUrl = normalizeBaseUrl(settings.apiBaseUrl);
  const uploadResponse = await uploadContent(apiBaseUrl, activeFile.name, uploadMarkdown);
  const templateName = readTemplateName(frontmatter, settings.templateFallback);
  const publishResponse = await createPublishJob(apiBaseUrl, uploadResponse.id, targetIds, templateName);

  const summary = [
    `Published ${activeFile.basename}`,
    `contentId=${uploadResponse.id}`,
    `jobId=${publishResponse.jobId}`,
    `targets=${targetIds.length}`,
  ].join(" | ");

  if (settings.showNotice) {
    new obsidian.Notice(summary, 8000);
  }

  return {
    contentId: uploadResponse.id,
    jobId: publishResponse.jobId,
    targetIds,
    templateName,
    file: activeFile.path,
    summary,
  };
}

function readFrontmatter(app, file) {
  const cache = app.metadataCache.getFileCache(file);
  return cache?.frontmatter ?? {};
}

function validatePostConfig(frontmatter) {
  if (frontmatter.post_publish !== true) {
    throw new Error("post_publish must be true");
  }

  if (!Array.isArray(frontmatter.post_targets) || frontmatter.post_targets.length === 0) {
    throw new Error("post_targets must be a non-empty list");
  }

  const invalid = frontmatter.post_targets.find((item) => typeof item !== "string" || item.trim() === "");
  if (invalid !== undefined) {
    throw new Error("post_targets must contain non-empty strings");
  }
}

async function readAliasMap(app, vaultRelativePath) {
  if (!vaultRelativePath || typeof vaultRelativePath !== "string") {
    throw new Error("aliasMappingPath is required");
  }

  const file = app.vault.getAbstractFileByPath(vaultRelativePath);
  if (!file) {
    throw new Error(`Alias mapping file not found: ${vaultRelativePath}`);
  }

  const raw = await app.vault.read(file);
  let parsed;
  try {
    parsed = JSON.parse(raw);
  } catch (error) {
    throw new Error(`Alias mapping JSON is invalid: ${error.message}`);
  }

  if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
    throw new Error("Alias mapping must be a JSON object");
  }

  return parsed;
}

function resolveTargetIds(targetAliases, aliasMap) {
  const targetIds = [];
  const missingAliases = [];

  for (const rawAlias of targetAliases) {
    const alias = String(rawAlias).trim();
    const mapped = aliasMap[alias];
    if (!Array.isArray(mapped) || mapped.length === 0) {
      missingAliases.push(alias);
      continue;
    }

    for (const rawTargetId of mapped) {
      const targetId = String(rawTargetId).trim();
      if (targetId !== "" && !targetIds.includes(targetId)) {
        targetIds.push(targetId);
      }
    }
  }

  if (missingAliases.length > 0) {
    throw new Error(`Missing target alias mapping: ${missingAliases.join(", ")}`);
  }

  if (targetIds.length === 0) {
    throw new Error("No target IDs resolved from post_targets");
  }

  return targetIds;
}

function buildUploadMarkdown(frontmatter, rawMarkdown, obsidian) {
  const uploadFrontmatter = { ...frontmatter };
  const postTitle = typeof frontmatter.post_title === "string" ? frontmatter.post_title.trim() : "";
  if (postTitle !== "") {
    uploadFrontmatter.title = postTitle;
  }

  delete uploadFrontmatter.post_publish;
  delete uploadFrontmatter.post_targets;
  delete uploadFrontmatter.post_template;

  const body = stripFrontmatter(rawMarkdown);
  const yaml = obsidian.stringifyYaml(uploadFrontmatter).trim();

  if (yaml === "{}") {
    return body.trimStart();
  }

  return `---\n${yaml}\n---\n\n${body.trimStart()}`;
}

function stripFrontmatter(markdown) {
  return String(markdown).replace(/^---\n[\s\S]*?\n---\n?/, "");
}

function normalizeBaseUrl(apiBaseUrl) {
  return String(apiBaseUrl || "").trim().replace(/\/+$/, "");
}

function readTemplateName(frontmatter, fallback) {
  const template = typeof frontmatter.post_template === "string" ? frontmatter.post_template.trim() : "";
  if (template !== "") {
    return template;
  }

  const normalizedFallback = String(fallback || "").trim();
  return normalizedFallback === "" ? "default" : normalizedFallback;
}

async function uploadContent(apiBaseUrl, fileName, markdown) {
  const formData = new FormData();
  formData.append("file", new Blob([markdown], { type: "text/markdown;charset=utf-8" }), fileName);

  const response = await fetch(`${apiBaseUrl}/contents/upload`, {
    method: "POST",
    body: formData,
  });

  return handleJsonResponse(response, "Upload content failed");
}

async function createPublishJob(apiBaseUrl, contentId, targetIds, templateName) {
  const response = await fetch(`${apiBaseUrl}/publish-jobs`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      contentId,
      targetIds,
      templateName,
    }),
  });

  return handleJsonResponse(response, "Create publish job failed");
}

async function handleJsonResponse(response, fallbackMessage) {
  const text = await response.text();
  let payload = {};
  if (text) {
    try {
      payload = JSON.parse(text);
    } catch {
      payload = {};
    }
  }

  if (!response.ok) {
    const message =
      typeof payload.message === "string" && payload.message.trim() !== ""
        ? payload.message.trim()
        : fallbackMessage;
    const code = typeof payload.code === "string" && payload.code.trim() !== "" ? payload.code.trim() : null;
    throw new Error(code ? `${message} (${code})` : message);
  }

  return payload;
}
