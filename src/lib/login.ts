const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export interface LoginErrors {
  email?: string;
  password?: string;
}

export interface LoginValues {
  email: string;
  password: string;
}

export function validateLogin(email: string, password: string): LoginErrors {
  return validateLoginValues({ email, password });
}

export function validateLoginValues({ email, password }: LoginValues): LoginErrors {
  const errors: LoginErrors = {};
  const normalizedEmail = email.trim();

  if (!normalizedEmail) {
    errors.email = 'Isi email.';
  } else if (normalizedEmail.length > 254 || !EMAIL_PATTERN.test(normalizedEmail)) {
    errors.email = 'Masukkan email yang valid.';
  }

  if (!password) {
    errors.password = 'Isi kata sandi.';
  } else if (password.length > 128) {
    errors.password = 'Kata sandi terlalu panjang.';
  }

  return errors;
}

export function hasErrors(errors: LoginErrors): boolean {
  return Object.keys(errors).length > 0;
}

export function safeNextPath(value: string | null): string {
  if (!value || !value.startsWith('/') || value.startsWith('//') || value.startsWith('/\\')) {
    return '/akun';
  }
  return value;
}
