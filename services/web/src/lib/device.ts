export const generateDeviceId = (): string => {
  // 1. Check if we already have a deviceId in storage
  const savedId = localStorage.getItem('device_id');

  if (savedId) {
    return savedId;
  }

  // 2. Generate a new unique ID
  // crypto.randomUUID() is the modern standard for generating UUID v4
  const newId = crypto.randomUUID();

  // 3. Save it for future visits
  localStorage.setItem('device_id', newId);

  return newId;
};
