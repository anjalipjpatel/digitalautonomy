import * as React from "react"
import { Link, type LinkProps } from "react-router-dom"
import { type VariantProps } from "class-variance-authority"
import { buttonVariants } from "@/components/ui/shadcn/button"
import { cn } from "@/lib/utils"


export interface LinkButtonProps extends LinkProps, VariantProps<typeof buttonVariants> {
    external?: boolean
}

export function ButtonLink({className, variant, size, to, external = false, children, ...props}: LinkButtonProps) {
  const styles = cn(buttonVariants({variant, size, className}))

  // Render using <a> for external links
  if (external || (typeof to === "string" && to.startsWith("http"))) {
    return (
      <a href={to.toString()} className={styles} target="_blank" rel="noopener noreferrer" {...(props as React.AnchorHTMLAttributes<HTMLAnchorElement>)} >
        {children}
      </a>
    )
  }

  // Render normal link otherwise
  return (
    <Link to={to} className={styles} {...props}>
      {children}
    </Link>
  )
}
