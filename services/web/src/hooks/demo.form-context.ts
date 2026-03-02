import { createContext, useContext } from 'react';

const FormContext = createContext<any>(null);
const FieldContext = createContext<any>(null);

export function useFormContext() {
  const ctx = useContext(FormContext);

  if (!ctx) {
    throw new Error('useFormContext must be used within a form provider');
  }

  return ctx;
}

export function useFieldContext<T = unknown>() {
  const ctx = useContext(FieldContext);

  if (!ctx) {
    throw new Error('useFieldContext must be used within a field provider');
  }

  return ctx as T;
}
