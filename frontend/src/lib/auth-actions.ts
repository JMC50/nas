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
