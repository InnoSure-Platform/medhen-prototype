// Role → permission model. The backend enforces RBAC authoritatively; this
// mirror lets the UI hide/disable actions the user can't perform and gives the
// proxy a defense-in-depth check. `primaryRole` is derived in the NextAuth jwt
// callback (see app/api/auth/[...nextauth]/route.ts).

export type Role = "customer" | "broker" | "staff" | "claims" | "admin" | "";

export type Permission =
  | "quote:create"
  | "policy:read"
  | "policy:read:own"
  | "policy:service"
  | "claim:create"
  | "claim:read"
  | "claim:settle"
  | "invoice:pay"
  | "book:read"
  | "commission:read"
  | "audit:read"
  | "kpi:read"
  | "underwriting:read"
  | "users:manage"
  | "product:manage"
  | "tenant:manage";

const CUSTOMER: Permission[] = ["quote:create", "policy:read:own", "claim:create", "invoice:pay"];
const BROKER: Permission[] = [...CUSTOMER, "book:read", "commission:read", "policy:service"];
const STAFF: Permission[] = ["policy:read", "policy:service", "claim:read", "audit:read", "kpi:read", "underwriting:read", "quote:create"];
const CLAIMS: Permission[] = [...STAFF, "claim:settle"];

// "*" grants everything (admin).
const PERMISSIONS: Record<Role, Permission[] | "*"> = {
  customer: CUSTOMER,
  broker: BROKER,
  staff: STAFF,
  claims: CLAIMS,
  admin: "*",
  "": [],
};

export function hasPermission(role: Role, perm: Permission): boolean {
  const set = PERMISSIONS[role] ?? [];
  return set === "*" || set.includes(perm);
}

/** Portal home for a role (used for post-login landing). */
export function homeForRole(role: Role): string {
  switch (role) {
    case "admin":
      return "/admin";
    case "staff":
    case "claims":
      return "/staff";
    case "broker":
      return "/broker";
    case "customer":
      return "/customer";
    default:
      return "/customer";
  }
}
