import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

/**
 * cn merges Tailwind class names, letting a later class win over an earlier one of
 * the same kind. Every shadcn component expects this helper under this name.
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
