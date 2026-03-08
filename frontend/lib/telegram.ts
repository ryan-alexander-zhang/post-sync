export type TelegramTargetConfig = {
  chatId?: string;
  topicId?: number;
  topicName?: string;
  disableNotification?: boolean;
  disableWebPagePreview?: boolean;
};

export function parseTelegramTargetConfig(raw: string): TelegramTargetConfig {
  if (!raw) {
    return {};
  }

  try {
    return JSON.parse(raw) as TelegramTargetConfig;
  } catch {
    return {};
  }
}
