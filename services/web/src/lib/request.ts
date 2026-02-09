import type { Metadata } from '@/types/response';
import { generateDeviceId } from './device';

export function parseBody<T = object>(body: T): T & { metadata: Metadata } {
  return {
    ...body,
    metadata: {
      deviceId: generateDeviceId(),
    },
  };
}
