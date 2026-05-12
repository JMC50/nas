import { useAuth } from "$lib/store/store";

interface LoginPayload {
  userId: string;
  password: string;
}

interface RegisterPayload {
  userId: string;
  username: string;
  password: string;
  krname: string;
}

interface AuthResult {
  success: boolean;
  message?: string;
}

function redirectHome() {
  window.location.replace(`${location.protocol}//${location.host}/`);
}

export async function signInLocal(payload: LoginPayload): Promise<AuthResult> {
  try {
    const response = await fetch("/server/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    const data = await response.json();
    if (!data.success) return { success: false, message: data.message ?? "Login failed." };
    useAuth.set({
      userId: data.user.userId,
      username: data.user.username,
      krname: data.user.krname ?? "",
      global_name: data.user.global_name,
      token: data.token,
    });
    redirectHome();
    return { success: true };
  } catch {
    return { success: false, message: "Network error. Please try again." };
  }
}

export async function signUpLocal(payload: RegisterPayload): Promise<AuthResult> {
  try {
    const response = await fetch("/server/auth/register", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    const data = await response.json();
    if (!data.success) return { success: false, message: data.message ?? "Registration failed." };
    useAuth.set({
      userId: data.user.userId,
      username: data.user.username,
      krname: payload.krname,
      global_name: data.user.global_name,
      token: data.token,
    });
    redirectHome();
    return { success: true };
  } catch {
    return { success: false, message: "Network error. Please try again." };
  }
}

export function discordUrl(): string {
  return (typeof process !== "undefined" ? process.env.LOGIN_URL : "") as string;
}

export function googleUrl(): string {
  const clientId = (typeof process !== "undefined" ? process.env.GOOGLE_CLIENT_ID : "") as string;
  const redirect = (typeof process !== "undefined" ? process.env.GOOGLE_REDIRECT_URI : "") as string;
  if (!clientId || !redirect) return "";
  const scope = encodeURIComponent("openid email profile");
  return `https://accounts.google.com/o/oauth2/v2/auth?response_type=code&client_id=${clientId}&redirect_uri=${encodeURIComponent(redirect)}&scope=${scope}`;
}
