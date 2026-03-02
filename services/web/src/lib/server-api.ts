import { createServerFn } from '@tanstack/react-start';
import axios, {
  type AxiosError,
  type AxiosRequestConfig,
  type AxiosRequestHeaders,
  type Method,
} from 'axios';

interface ServerErrorShape {
  success?: boolean;
  message: string;
}

type ApiAdapterInput = {
  method: Method;
  url: string;
  data?: unknown;
  accessToken?: string;
  headers?: AxiosRequestHeaders;
};

function normalizePath(path: string): string {
  return path.startsWith('/') ? path : `/${path}`;
}

function stripTrailingSlash(value: string): string {
  return value.endsWith('/') ? value.slice(0, -1) : value;
}

function resolveApiBaseUrl(): string {

  return (
    process.env.API_BASE_URL || 'localhost:8906/api/v1'
  );
}

function resolveRequestUrl(url: string): string {
  const baseUrl = resolveApiBaseUrl();
  const path = normalizePath(url);

  return `${stripTrailingSlash(baseUrl)}${path}`;
}

function buildHeaders(input: ApiAdapterInput): AxiosRequestHeaders | undefined {
  if (!input.accessToken) {
    return input.headers;
  }

  return {
    ...input.headers,
    Authorization: `Bearer ${input.accessToken}`,
  };
}

export async function apiAdapter<
  TSuccess,
  TError extends ServerErrorShape = ServerErrorShape,
>(input: ApiAdapterInput): Promise<TSuccess | TError> {
  const config: AxiosRequestConfig = {
    method: input.method,
    url: resolveRequestUrl(input.url),
    data: input.data,
    headers: buildHeaders(input),
    timeout: 5000,
  };

  try {
    const res = await axios.request<TSuccess>(config);
    return res.data;
  } catch (error) {
    const err = error as AxiosError<TError>;

    if (err.response?.data) {
      return err.response.data;
    }

    return {
      success: false,
      message: err.message || 'Request failed',
    } as TError;
  }
}

export const getBaseApiUrlFn = createServerFn({ method: 'GET' }).handler(
  async () => {
    return process.env.API_BASE_URL || 'not found';
  },
);
