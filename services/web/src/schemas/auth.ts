import { z } from 'zod'

export const signinSchema = z.object({
  email: z.email('Invalid email address'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
})
export type LoginSchemaType = z.infer<typeof signinSchema>
export const signupSchema = z.object({
  firstName: z.string().min(2, 'First name is required'),
  lastName: z.string().min(2, 'Last name is required'),
  email: z.email('Invalid email address'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
  confirmPassword: z.string(),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
})

export type SignupSchemaType = z.infer<typeof signupSchema>

export type SigninRequest = LoginSchemaType & {
  metadata: {
    deviceId: string;
  }
}

export type SignupRequest = SignupSchemaType & {
  metadata: {
    deviceId: string;
  }
}
