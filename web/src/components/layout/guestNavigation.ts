/**
 * The guest navigation, defined once.
 *
 * The bottom bar on a phone and the top nav on a desktop render this same list, and
 * `/mehr` renders the overflow half of it. Two hand-written lists would drift apart on
 * the first entry somebody adds, and the entry they forgot would be the one an
 * unconfident guest was hunting for.
 */

import {
  CalendarDays,
  ChevronsRight,
  Gift,
  HelpCircle,
  Home,
  MapPin,
  MoreHorizontal,
  Phone,
  Shield,
  Shirt,
  Table2,
  Image as ImageIcon,
} from "lucide-react";

import type { Flags } from "@/lib/api/dto";
import { navLabels } from "@/lib/labels";

export interface NavEntry {
  to: string;
  label: string;
  icon: typeof Home;
}

/**
 * The five bottom-bar entries.
 *
 * Antwort sits in the bar rather than on `/mehr` because it is the one thing we
 * actually need a guest to do; its label follows whether the household has answered.
 */
export function barEntries(hasAnswered: boolean): NavEntry[] {
  return [
    { to: "/start", label: navLabels.start, icon: Home },
    { to: "/ablauf", label: navLabels.schedule, icon: CalendarDays },
    { to: "/location", label: navLabels.location, icon: MapPin },
    { to: "/zusagen", label: hasAnswered ? navLabels.rsvpAnswered : navLabels.rsvp, icon: ChevronsRight },
    { to: "/mehr", label: navLabels.more, icon: MoreHorizontal },
  ];
}

/**
 * Everything else, as the list `/mehr` renders.
 *
 * Flag-gated entries are **absent** rather than disabled: an unpublished seating plan
 * is not "coming soon", it is nothing a guest should be thinking about yet.
 */
export function moreEntries(flags: Flags): NavEntry[] {
  return [
    { to: "/dresscode", label: navLabels.dresscode, icon: Shirt },
    { to: "/geschenke", label: navLabels.gifts, icon: Gift },
    { to: "/faq", label: navLabels.faq, icon: HelpCircle },
    { to: "/kontakt", label: navLabels.contact, icon: Phone },
    { to: "/datenschutz", label: navLabels.privacy, icon: Shield },
    ...(flags.seating_published ? [{ to: "/sitzplan", label: navLabels.seating, icon: Table2 }] : []),
    ...(flags.gallery_visible ? [{ to: "/galerie", label: navLabels.gallery, icon: ImageIcon }] : []),
  ];
}
