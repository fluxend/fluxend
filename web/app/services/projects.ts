import { get, type APIRequestOptions } from "~/tools/fetch";

export type Project = {
  uuid: string;
  name: string;
  status: "active" | "inactive";
  dbName: string;
  organizationUuid: string;
  createdAt: string;
  updatedAt: string;
};

// Project tokens are valid for 24 hours; refresh well before that
const projectTokenCache = new Map<string, { token: string; fetchedAt: number }>();
const PROJECT_TOKEN_TTL_MS = 60 * 60 * 1000;

export const createProjectsService = (authToken: string) => {
  const getProjectDetails = async (projectId: string): Promise<Project> => {
    try {
      const fetchOptions: APIRequestOptions = {
        headers: {
          "Content-Type": "application/json",
          ...(authToken && { Authorization: `Bearer ${authToken}` }),
        },
      };

      const response = await get(`/projects/${projectId}`, fetchOptions);
      const result = await response.json();

      if (result.success && result.content) {
        return result.content;
      } else {
        throw new Error(result.errors?.join(", ") || "Unknown error");
      }
    } catch (error) {
      // Fallback to mock data for development
      console.warn("Failed to fetch health status, using mock data:", error);
      throw new Error("Failed to fetch project details");
    }
  };

  const getProjectToken = async (projectId: string): Promise<string> => {
    // Key by session token too: a project token embeds the requesting user's
    // database role, so it must never be reused across sessions
    const cacheKey = `${authToken}:${projectId}`;
    const cached = projectTokenCache.get(cacheKey);
    if (cached && Date.now() - cached.fetchedAt < PROJECT_TOKEN_TTL_MS) {
      return cached.token;
    }

    const fetchOptions: APIRequestOptions = {
      headers: {
        "Content-Type": "application/json",
        ...(authToken && { Authorization: `Bearer ${authToken}` }),
      },
    };

    const response = await get(`/projects/${projectId}/token`, fetchOptions);
    const result = await response.json();

    if (!result.success || !result.content?.token) {
      throw new Error(result.errors?.join(", ") || "Failed to fetch project token");
    }

    projectTokenCache.set(cacheKey, {
      token: result.content.token,
      fetchedAt: Date.now(),
    });

    return result.content.token;
  };

  return {
    getProjectDetails,
    getProjectToken,
  };
};

export type ProjectService = ReturnType<typeof createProjectsService>;
