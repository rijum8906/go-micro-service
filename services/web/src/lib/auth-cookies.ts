import {
  deleteCookie,
  getCookie,
  setCookie,
} from '@tanstack/react-start/server';

const TOKEN_COOKIE_NAME = 'access_token';

export function getAuthToken() {
  return getCookie(TOKEN_COOKIE_NAME);
}

export function setAuthToken(token: string) {
  setCookie(TOKEN_COOKIE_NAME, token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    path: '/',
    maxAge: 60 * 60 * 24 * 7, // 1 week
  });
}

export function clearAuthToken() {
  deleteCookie(TOKEN_COOKIE_NAME, {
    path: '/',
  });
}
