"use client";

import { LogOut, ShieldCheck, UserRound } from "lucide-react";
import { useTranslations } from "next-intl";
import { useAuth } from "@/components/providers";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

export function UserMenu() {
  const { user, role, logout, enrollMfa, accountUrl } = useAuth();
  const t = useTranslations();
  if (!user) return null;
  const initial = user.name.slice(0, 1).toUpperCase();

  return (
    <DropdownMenu>
      <DropdownMenuTrigger className="flex items-center gap-2 rounded-full p-0.5 outline-none focus-visible:ring-2 focus-visible:ring-ring">
        <Avatar>
          <AvatarFallback>{initial}</AvatarFallback>
        </Avatar>
        <span className="hidden max-w-[9rem] truncate pr-1 text-sm font-medium text-fg sm:inline">{user.name}</span>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="min-w-[14rem]">
        <DropdownMenuLabel>
          <span className="block truncate text-sm font-semibold normal-case tracking-normal text-fg">{user.name}</span>
          <span className="block text-xs font-normal text-fg-muted">
            {role ? t(`roles.${role}` as never) : "—"}
          </span>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        {accountUrl && (
          <DropdownMenuItem asChild>
            <a href={accountUrl} target="_blank" rel="noopener noreferrer">
              <UserRound /> {t("auth.account")}
            </a>
          </DropdownMenuItem>
        )}
        <DropdownMenuItem onClick={() => enrollMfa()}>
          <ShieldCheck /> {t("auth.enable2fa")}
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={logout} className="text-danger">
          <LogOut /> {t("common.signOut")}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
