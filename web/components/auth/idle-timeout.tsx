"use client";

import * as React from "react";
import { useTranslations } from "next-intl";
import { TimerReset } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";

const IDLE_MS = 15 * 60 * 1000; // sign out after 15 min idle
const WARN_MS = 60 * 1000; // warn 60s before

/**
 * Signs the user out after a period of inactivity, with a warning + countdown.
 * Any real activity (pointer/keyboard/scroll) resets the timer.
 */
export function IdleTimeout({ onTimeout }: { onTimeout: () => void }) {
  const t = useTranslations("auth");
  const lastActivity = React.useRef(Date.now());
  const [warnOpen, setWarnOpen] = React.useState(false);
  const [secondsLeft, setSecondsLeft] = React.useState(Math.ceil(WARN_MS / 1000));

  const reset = React.useCallback(() => {
    lastActivity.current = Date.now();
    setWarnOpen(false);
  }, []);

  React.useEffect(() => {
    const events = ["mousemove", "mousedown", "keydown", "scroll", "touchstart", "visibilitychange"];
    const onActivity = () => {
      // Ignore activity while the warning is up so the countdown is deliberate.
      if (!warnOpen) lastActivity.current = Date.now();
    };
    events.forEach((e) => window.addEventListener(e, onActivity, { passive: true }));
    return () => events.forEach((e) => window.removeEventListener(e, onActivity));
  }, [warnOpen]);

  React.useEffect(() => {
    const tick = setInterval(() => {
      const idle = Date.now() - lastActivity.current;
      if (idle >= IDLE_MS) {
        onTimeout();
      } else if (idle >= IDLE_MS - WARN_MS) {
        setWarnOpen(true);
        setSecondsLeft(Math.max(0, Math.ceil((IDLE_MS - idle) / 1000)));
      }
    }, 1000);
    return () => clearInterval(tick);
  }, [onTimeout]);

  return (
    <Dialog open={warnOpen} onOpenChange={(o) => !o && reset()}>
      <DialogContent hideClose>
        <DialogHeader>
          <span className="grid size-11 place-items-center rounded-xl border border-warning/20 bg-warning-subtle text-warning-fg">
            <TimerReset className="size-5" />
          </span>
          <DialogTitle>{t("idleTitle")}</DialogTitle>
          <DialogDescription>{t("idleBody", { seconds: secondsLeft })}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="ghost" onClick={onTimeout}>{t("signOutNow")}</Button>
          <Button onClick={reset}>{t("stay")}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
