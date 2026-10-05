import type { PasswordPolicyCheck } from '../../components/PasswordField';

/** Datos de la cuenta con los que la contraseña nueva no puede coincidir (FR-010). */
export interface PasswordPolicyContext {
  firstName?: string | null;
  lastName?: string | null;
  email?: string | null;
}

/** Longitud mínima de la política vigente (contrato `ChangePasswordInput`). */
export const PASSWORD_MIN_LENGTH = 8;
/** Longitud máxima de la política vigente (contrato `ChangePasswordInput`). */
export const PASSWORD_MAX_LENGTH = 64;

const UPPER = /\p{Lu}/u;
const LOWER = /\p{Ll}/u;
const DIGIT = /\p{Nd}/u;
const LETTER_DIGIT_OR_SPACE = /[\p{L}\p{N}\s]/u;

/** Comparación acordada en la spec (Q5/R4): trim + minúsculas. */
function normalize(value: string): string {
  return value.trim().toLowerCase();
}

/** Recorre por puntos de código (equivalente a runas) y evalúa el predicado. */
function someCodePoint(value: string, predicate: (char: string) => boolean): boolean {
  for (const char of value) {
    if (predicate(char)) {
      return true;
    }
  }
  return false;
}

/**
 * Evalúa la política FR-010 en vivo (ux.md §3.5/§3.8) para el checklist del
 * `PasswordField`. Devuelve un requisito por regla con su estado.
 */
export function evaluatePasswordPolicy(
  password: string,
  context: PasswordPolicyContext = {},
): PasswordPolicyCheck[] {
  const length = Array.from(password).length;
  const personalData = [context.firstName, context.lastName, context.email]
    .filter((value): value is string => typeof value === 'string' && value.trim() !== '')
    .map((value) => normalize(value));
  const equalsPersonalData = password.trim() !== '' && personalData.includes(normalize(password));

  return [
    {
      id: 'min',
      label: `Al menos ${PASSWORD_MIN_LENGTH} caracteres`,
      met: length >= PASSWORD_MIN_LENGTH,
    },
    {
      id: 'max',
      label: `Como máximo ${PASSWORD_MAX_LENGTH} caracteres`,
      met: length > 0 && length <= PASSWORD_MAX_LENGTH,
    },
    { id: 'upper', label: 'Una letra mayúscula', met: UPPER.test(password) },
    { id: 'lower', label: 'Una letra minúscula', met: LOWER.test(password) },
    { id: 'digit', label: 'Un número', met: DIGIT.test(password) },
    {
      id: 'special',
      label: 'Un carácter especial',
      met: someCodePoint(password, (char) => !LETTER_DIGIT_OR_SPACE.test(char)),
    },
    {
      id: 'personal',
      label: 'Distinta del nombre, los apellidos y el correo',
      met: !equalsPersonalData,
    },
  ];
}

/** `true` solo cuando la contraseña cumple todos los requisitos (no vacía). */
export function isPasswordPolicyMet(checks: readonly PasswordPolicyCheck[]): boolean {
  return checks.length > 0 && checks.every((check) => check.met);
}
