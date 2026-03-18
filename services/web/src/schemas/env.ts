import z from "zod";

export const envSchema = z.object({
  NODE_ENV: z.enum(['development', 'test', 'production']).default('production'),
  PORT: z.coerce.number().default(3333),

  // API 
  API_BASE_URL: z.url().default('http://localhost:8906/api/v1'),
});

export type EnvSchemaType = z.infer<typeof envSchema>;
