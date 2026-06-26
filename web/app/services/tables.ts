import type { APIResponse } from "~/lib/types";
import { getTypedResponseData } from "~/lib/utils";
import type { Table } from "~/routes/tables/table-list";
import {get, post, del, put, type APIRequestOptions, patch} from "~/tools/fetch";

export function createTablesService(authToken: string) {
  const getAllTables = async (projectId: string) => {
    const fetchOptions: RequestInit = {
      headers: {
        "X-Project": projectId,
        "Content-Type": "application/json",
        Authorization: `Bearer ${authToken}`,
      },
    };

    const response = await get("/tables", fetchOptions);
    const data = await getTypedResponseData<APIResponse<any>>(response);

    return data;
  };

  const getTableColumns = async (projectId: string, collectionName: string) => {
    const fetchOptions: RequestInit = {
      headers: {
        "X-Project": projectId,
        "Content-Type": "application/json",
        Authorization: `Bearer ${authToken}`,
      },
    };

    return get(`/tables/public.${collectionName}/columns`, fetchOptions);
  };

  const getTableRows = async (
    projectId: string,
    collectionName: string,
    options?: {
      headers?: HeadersInit;
      params?: Record<string, any>;
      baseUrl?: string;
    }
  ) => {
    const fetchOptions: APIRequestOptions = {
      headers: {
        "X-Project": projectId,
        "Content-Type": "application/json",
        Authorization: `Bearer ${authToken}`,
        ...options?.headers,
      },
      params: options?.params,
      baseUrl: options?.baseUrl,
    };

    return get(collectionName, fetchOptions);
  };

  const createTable = async (projectId: string, data: any) => {
    const fetchOptions: RequestInit = {
      headers: {
        "X-Project": projectId,
        "Content-Type": "application/json",
        Authorization: `Bearer ${authToken}`,
      },
    };

    return post("/tables", data, fetchOptions);
  };

  const deleteTable = async (projectId: string, tableName: string) => {
    const fetchOptions: RequestInit = {
      headers: {
        "X-Project": projectId,
        "Content-Type": "application/json",
        Authorization: `Bearer ${authToken}`,
      },
    };

    const response = await del(`/tables/public.${tableName}`, fetchOptions);
    const data = await getTypedResponseData<APIResponse<null>>(response);
    return data;
  };

  const createTableColumns = async (projectId: string, tableName: string, data: any) => {
    const fetchOptions: RequestInit = {
      headers: {
        "X-Project": projectId,
        "Content-Type": "application/json",
        Authorization: `Bearer ${authToken}`,
      },
    };

    return post(`/tables/public.${tableName}/columns`, data, fetchOptions);
  };

  const updateTableColumns = async (projectId: string, tableName: string, data: any) => {
    const fetchOptions: RequestInit = {
      headers: {
        "X-Project": projectId,
        "Content-Type": "application/json",
        Authorization: `Bearer ${authToken}`,
      },
    };

    return patch(`/tables/public.${tableName}/columns`, data, fetchOptions);
  };

  const createTableRow = async (projectId: string, tableId: string, data: any) => {
    const fetchOptions: APIRequestOptions = {
      headers: {
        "X-Project": projectId,
        "Content-Type": "application/json",
        Authorization: `Bearer ${authToken}`,
        "Prefer": "return=representation",
      },
    };

    return post(`/tables/public.${tableId}/rows`, data, fetchOptions);
  };

  const updateTableRow = async (
    projectId: string,
    tableId: string,
    rowId: string,
    data: any,
  ) => {
    const fetchOptions: APIRequestOptions = {
      headers: {
        "X-Project": projectId,
        "Content-Type": "application/json",
        Authorization: `Bearer ${authToken}`,
        "Prefer": "return=representation",
      },
      params: {
        id: `eq.${rowId}`,
      },
    };

    return patch(`/tables/public.${tableId}/rows`, data, fetchOptions);
  };

  const deleteTableRows = async (
    projectId: string,
    tableId: string,
    rowIds: string[],
  ) => {
    if (!rowIds || rowIds.length === 0) {
      throw new Error("No row IDs provided for deletion");
    }

    const escapedIds = rowIds
      .map((id) => String(id).replace(/[^a-zA-Z0-9-_]/g, ""))
      .filter(Boolean);

    if (escapedIds.length === 0) {
      throw new Error("No valid row IDs provided for deletion");
    }

    const fetchOptions: APIRequestOptions = {
      headers: {
        "X-Project": projectId,
        "Content-Type": "application/json",
        Authorization: `Bearer ${authToken}`,
        "Prefer": "return=representation",
      },
      params: {
        id: `in.(${escapedIds.join(",")})`,
      },
    };

    return del(`/tables/public.${tableId}/rows`, fetchOptions);
  };

  return {
    getAllTables,
    getTableColumns,
    getTableRows,
    createTable,
    deleteTable,
    createTableColumns,
    updateTableColumns,
    createTableRow,
    updateTableRow,
    deleteTableRows,
  };
}

export type TablesService = ReturnType<typeof createTablesService>;
