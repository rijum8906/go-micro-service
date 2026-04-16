import { z } from 'zod';

export const signinSchema = z.object({
  email: z.email('Invalid email address'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
});
export type SigninSchemaType = z.infer<typeof signinSchema>;
export const signupBaseSchema = z.object({
  firstName: z.string().min(2, 'First name is required'),
  lastName: z.string().min(2, 'Last name is required'),
  email: z.email('Invalid email address'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
  confirmPassword: z.string(),
});

export const signupSchema = signupBaseSchema.refine(
  (data) => data.password === data.confirmPassword,
  { message: "Passwords don't match", path: ['confirmPassword'] },
);

export type SignupSchemaType = z.infer<typeof signupSchema>;

export const updateProfileSchema = z.object({
  firstName: z.string().optional(),
  lastName: z.string().optional(),
  displayName: z.string().optional(),
  avatar: z.file().optional(),
});
export type UpdateProfileSchemaType = z.infer<typeof updateProfileSchema>;

export const createProfileSchema = z.object({
  firstName: z.string().min(2, 'First name is required'),
  lastName: z.string().min(2, 'Last name is required'),
  displayName: z.string().optional(),
  avatar: z.file().optional(),
});
export type CreateProfileSchemaType = z.infer<typeof createProfileSchema>;


export const changePasswordSchema = z
  .object({
    currentPassword: z.string().min(8, 'Password must be at least 8 characters'),
    newPassword: z.string().min(8, 'Password must be at least 8 characters'),
    confirmPassword: z.string(),
  })
  .refine((data) => data.newPassword === data.confirmPassword, {
    message: "Passwords don't match",
    path: ['confirmPassword'],
  });
export type ChangePasswordSchemaType = z.infer<typeof changePasswordSchema>;

export const requestPasswordResetSchema = z.object({
  email: z.email('Invalid email address'),
});
export type RequestPasswordResetSchemaType = z.infer<typeof requestPasswordResetSchema>;

export const resetPasswordSchema = z
  .object({
    newPassword: z.string().min(8, 'Password must be at least 8 characters'),
    confirmPassword: z.string(),
  })
  .refine((data) => data.newPassword === data.confirmPassword, {
    message: "Passwords don't match",
    path: ['confirmPassword'],
  });
export type ResetPasswordSchemaType = z.infer<typeof resetPasswordSchema>;
