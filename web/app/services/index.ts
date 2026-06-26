import { createTablesService } from "./tables";
import { createDashboardService } from "./dashboard";
import { createOpenAPIService } from "./openapi";
import { createProjectsService } from "./projects";
import { createUserService } from "./user";
import { createLogsService } from "./logs";
import { createSettingsService } from "./settings";
import { createStorageService } from "./storage";
import { createWebhooksService } from "./webhooks";

export function initializeServices(authToken: string) {
  return {
    tables: createTablesService(authToken),
    dashboard: createDashboardService(authToken),
    openapi: createOpenAPIService(authToken),
    user: createUserService(authToken),
    projects: createProjectsService(authToken),
    logs: createLogsService(authToken),
    settings: createSettingsService(authToken),
    storage: createStorageService(authToken),
    webhooks: createWebhooksService(authToken),
  };
}

export type Services = ReturnType<typeof initializeServices>;
