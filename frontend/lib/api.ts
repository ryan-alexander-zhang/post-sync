import "server-only";

import {
  ChannelAccount,
  ChannelTarget,
  Content,
  DeliveryTask,
  PublishJob,
} from "@/lib/types";

const API_BASE =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api/v1";

async function request<T>(path: string): Promise<T> {
  try {
    const response = await fetch(`${API_BASE}${path}`, { cache: "no-store" });
    if (!response.ok) {
      throw new Error(`API request failed: ${path}`);
    }

    return response.json() as Promise<T>;
  } catch {
    throw new Error(`API unavailable: ${path}`);
  }
}

export async function getContents() {
  try {
    return await request<{ items: Content[] }>("/contents");
  } catch {
    return { items: [] };
  }
}

export async function getContent(id: string) {
  return request<Content>(`/contents/${id}`);
}

export async function getChannelAccounts() {
  try {
    return await request<{ items: ChannelAccount[] }>("/channel-accounts");
  } catch {
    return { items: [] };
  }
}

export async function getChannelTargets() {
  try {
    return await request<{ items: ChannelTarget[] }>("/channel-targets");
  } catch {
    return { items: [] };
  }
}

export async function getPublishJobs() {
  try {
    return await request<{ items: PublishJob[] }>("/publish-jobs");
  } catch {
    return { items: [] };
  }
}

export async function getPublishJob(id: string) {
  return request<{ job: PublishJob; deliveries: DeliveryTask[] }>(
    `/publish-jobs/${id}`,
  );
}

export async function getSystemInfo() {
  try {
    return await request<{ database: string; status: string }>("/system/info");
  } catch {
    return { database: "unavailable", status: "offline" };
  }
}
