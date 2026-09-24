import JSONBigFactory from "json-bigint";
import type { JsonId } from "../types/domain";

const jsonBig = JSONBigFactory({ useNativeBigInt: true });

export function parseJson<T>(value: string): T {
  return jsonBig.parse(value) as T;
}

export function stringifyJson(value: unknown): string {
  return jsonBig.stringify(value);
}

export function normalizeId(value: JsonId): string {
  return String(value);
}

export function backendId(value: string): bigint {
  if (!/^\d+$/.test(value)) {
    throw new Error("Invalid backend ID");
  }
  return BigInt(value);
}

