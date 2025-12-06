import { verify } from "jsonwebtoken";

export default () => {
  return async (ctx, next) => {
    const authHeader = ctx.request.headers.authorization;
    if (!authHeader) {
      return ctx.unauthorized("Token missing");
    }

    const token = authHeader.split(" ")[1];
    if (!token) {
      return ctx.unauthorized("Token missing");
    }

    try {
      const decoded = verify(token, process.env.JWT_SECRET) as {
        user_id: string;
      };
      ctx.state.user_id = decoded.user_id;
      await next();
    } catch (error) {
      console.log("Error verifying token", error);
      return ctx.unauthorized("Invalid token");
    }
  };
};
