export type Content = {
  id: string;
  sourceFilename: string;
  title: string;
  bodyHash: string;
  bodyMarkdown: string;
  bodyPlain: string;
  frontmatterJson: string;
  createdAt: string;
};

export type ChannelAccount = {
  id: string;
  channelType: string;
  name: string;
  enabled: boolean;
  secretRef: string;
  configJson: string;
  createdAt: string;
  updatedAt: string;
};

export type ChannelTarget = {
  id: string;
  channelAccountId: string;
  targetType: string;
  targetKey: string;
  targetName: string;
  enabled: boolean;
  configJson: string;
  createdAt: string;
  updatedAt: string;
};

export type PublishJob = {
  id: string;
  contentId: string;
  requestId: string;
  triggerSource: string;
  status: string;
  totalDeliveries: number;
  successCount: number;
  failedCount: number;
  skippedCount: number;
  createdAt: string;
  startedAt?: string;
  finishedAt?: string;
};

export type DeliveryTask = {
  id: string;
  publishJobId: string;
  channelAccountId: string;
  channelTargetId: string;
  channelType: string;
  targetKey: string;
  status: string;
  attemptCount: number;
  renderedBody: string;
  externalMessageId: string;
  errorCode: string;
  errorMessage: string;
  createdAt: string;
  startedAt?: string;
  finishedAt?: string;
};
