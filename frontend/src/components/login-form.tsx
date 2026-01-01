"use client";

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field"
import { Input, PasswordInputWithReveal } from "@/components/ui/input"
import { useActionState } from "react"
import { loginAction } from "@/app/login/actions"
import { useFormStatus } from "react-dom";
import { Spinner } from "./ui/spinner";
import AlertCard from "./ui/alert-card";

function SubmitButton() {
	const { pending } = useFormStatus();

	return <Button type="submit" disabled={pending}>
	Login
	{pending && <Spinner />}
	</Button>
}

export function LoginForm({
  className,
  ...props
}: React.ComponentProps<"div">) {
	const [state, action] = useActionState(loginAction, {});

  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        <CardHeader>
          <CardTitle>Login to your account</CardTitle>
          <CardDescription>
            Enter your email below to login to your account
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form action={action}>
            <FieldGroup>
              <Field>
                <FieldLabel htmlFor="email">Email</FieldLabel>
                <Input
                  id="email"
                  type="email"
									name="email"
                  placeholder="m@example.com"
                  required
                />
              </Field>
              <Field>
                <div className="flex items-center">
                  <FieldLabel htmlFor="password">Password</FieldLabel>
                </div>
                <PasswordInputWithReveal id="password" name="password" required />
              </Field>
							{state?.error && <AlertCard variant="error" title="Login failed" description={state.error} />}
              <Field>
								<SubmitButton />
                <FieldDescription className="text-center">
                  Don&apos;t have an account? <a href="#">Sign up</a>
                </FieldDescription>
              </Field>
            </FieldGroup>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
