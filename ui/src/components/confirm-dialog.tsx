import { Dialog, DialogContent } from "@/components/ui/dialog";
import type { DialogProps } from "@/hooks/use-dialog";
import type { PropsWithChildren } from "react";

export function ConfirmDialog({
  dialogProps,
  children,
}: PropsWithChildren<{
  dialogProps: DialogProps;
}>) {
  return (
    <Dialog {...dialogProps}>
      {/* This will contain the open and onOpenChange props */}
      <DialogContent>{children}</DialogContent>
    </Dialog>
  );
}
