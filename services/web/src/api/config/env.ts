import { envSchema, EnvSchemaType } from "@/schemas/env";
import z from "zod";

export var ENV: EnvSchemaType | null = null;

export function initEnv() {
  const parsedRes = z.safeParse(envSchema, process.env)
  if (!parsedRes.success) {
    throw new Error(parsedRes.error.message);
  }

  ENV = parsedRes.data;
  return ENV;
}
