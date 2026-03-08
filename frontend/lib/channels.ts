import { ChannelAccount, ChannelTarget } from "@/lib/types";

export type SupportedChannelType = "telegram" | "feishu" | "personal_feishu";

export type TelegramTargetConfig = {
  chatId?: string;
  topicId?: number;
  topicName?: string;
  disableNotification?: boolean;
  disableWebPagePreview?: boolean;
};

export type FeishuAccountConfig = {
  appIdEnv?: string;
  tokenEnv?: string;
  baseUrl?: string;
  signSecretRef?: string;
};

export type FeishuTargetConfig = {
  receiveIdType?: string;
  chatId?: string;
  webhookEnvRef?: string;
};

export const CHANNEL_OPTIONS: Array<{
  value: SupportedChannelType;
  label: string;
  description: string;
  defaultSecretRef: string;
}> = [
  {
    value: "telegram",
    label: "Telegram",
    description: "Use a bot token from the current environment.",
    defaultSecretRef: "TELEGRAM_BOT_TOKEN",
  },
  {
    value: "feishu",
    label: "Enterprise Feishu",
    description: "Use app credentials to fetch tenant access tokens.",
    defaultSecretRef: "FEISHU_APP_SECRET",
  },
  {
    value: "personal_feishu",
    label: "Personal Feishu",
    description: "Use environment refs for webhook URL and sign secret.",
    defaultSecretRef: "PERSONAL_FEISHU_WEBHOOK_URL",
  },
];

export function parseTelegramTargetConfig(raw: string): TelegramTargetConfig {
  return parseConfig<TelegramTargetConfig>(raw);
}

export function parseFeishuAccountConfig(raw: string): FeishuAccountConfig {
  return parseConfig<FeishuAccountConfig>(raw);
}

export function parseFeishuTargetConfig(raw: string): FeishuTargetConfig {
  return parseConfig<FeishuTargetConfig>(raw);
}

export function buildChannelAccountPayload(channelType: SupportedChannelType, formData: FormData) {
  if (channelType === "feishu") {
    return {
      channelType,
      name: formData.get("name"),
      secretRef: formData.get("secretRef"),
      config: {
        appIdEnv: formData.get("appIdEnv"),
        tokenEnv: formData.get("tokenEnv"),
        baseUrl: formData.get("baseUrl"),
      },
    };
  }

  if (channelType === "personal_feishu") {
    return {
      channelType,
      name: formData.get("name"),
      secretRef: formData.get("secretRef"),
      config: {
        signSecretRef: formData.get("signSecretRef"),
      },
    };
  }

  return {
    channelType,
    name: formData.get("name"),
    secretRef: formData.get("secretRef"),
    config: {},
  };
}

export function buildChannelTargetPayload(account: ChannelAccount | undefined, formData: FormData) {
  if (account?.channelType === "feishu") {
    return {
      channelAccountId: formData.get("channelAccountId"),
      targetType: "feishu_chat",
      targetKey: formData.get("targetKey"),
      targetName: formData.get("targetName"),
      config: {
        receiveIdType: "chat_id",
        chatId: formData.get("targetKey"),
      },
    };
  }

  if (account?.channelType === "personal_feishu") {
    return {
      channelAccountId: formData.get("channelAccountId"),
      targetType: "personal_feishu_webhook",
      targetKey: "",
      targetName: formData.get("targetName"),
      config: {},
    };
  }

  return {
    channelAccountId: formData.get("channelAccountId"),
    targetType:
      String(formData.get("topicId") || "").trim() === ""
        ? "telegram_group"
        : "telegram_topic",
    targetKey: formData.get("targetKey"),
    targetName: formData.get("targetName"),
    config: {
      chatId: formData.get("targetKey"),
      topicId: formData.get("topicId"),
      topicName: formData.get("topicName"),
      disableNotification: false,
      disableWebPagePreview: false,
    },
  };
}

export function describeTarget(target: ChannelTarget) {
  if (target.targetType === "feishu_chat") {
    const feishu = parseFeishuTargetConfig(target.configJson);
    return {
      channelLabel: "Enterprise Feishu",
      primary: feishu.chatId || target.targetKey,
      secondary: "Chat",
    };
  }

  if (target.targetType === "personal_feishu_webhook") {
    return {
      channelLabel: "Personal Feishu",
      primary: "Webhook bot",
      secondary: "Webhook",
    };
  }

  const telegram = parseTelegramTargetConfig(target.configJson);
  return {
    channelLabel: "Telegram",
    primary: telegram.chatId || target.targetKey,
    secondary:
      target.targetType === "telegram_topic"
        ? `${telegram.topicName || "Topic"} (${telegram.topicId ?? "-"})`
        : "Group root",
  };
}

function parseConfig<T>(raw: string): T {
  if (!raw) {
    return {} as T;
  }

  try {
    return JSON.parse(raw) as T;
  } catch {
    return {} as T;
  }
}
