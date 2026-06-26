import { get, post, del } from "~/tools/fetch";

export interface Webhook {
  uuid: string;
  tableName: string;
  url: string;
  events: string[];
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface CreateWebhookDto {
  url: string;
  events: string[];
  is_active: boolean;
}

export function createWebhooksService(authToken: string) {
  const listWebhooks = async (projectId: string, tableName: string) => {
    const fetchOptions: RequestInit = {
      headers: {
        "X-Project": projectId,
        "Content-Type": "application/json",
        Authorization: `Bearer ${authToken}`,
      },
    };

    return get(`/tables/public.${tableName}/webhooks`, fetchOptions);
  };

  const createWebhook = async (
    projectId: string,
    tableName: string,
    data: CreateWebhookDto
  ) => {
    const fetchOptions: RequestInit = {
      headers: {
        "X-Project": projectId,
        "Content-Type": "application/json",
        Authorization: `Bearer ${authToken}`,
      },
    };

    return post(`/tables/public.${tableName}/webhooks`, data, fetchOptions);
  };

  const deleteWebhook = async (
    projectId: string,
    tableName: string,
    webhookUUID: string
  ) => {
    const fetchOptions: RequestInit = {
      headers: {
        "X-Project": projectId,
        "Content-Type": "application/json",
        Authorization: `Bearer ${authToken}`,
      },
    };

    return del(
      `/tables/public.${tableName}/webhooks/${webhookUUID}`,
      fetchOptions
    );
  };

  return { listWebhooks, createWebhook, deleteWebhook };
}

export type WebhooksService = ReturnType<typeof createWebhooksService>;
