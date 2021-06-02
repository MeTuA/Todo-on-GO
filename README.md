# Todo-on-GO
Simple implementation of todo app on GO

`Login` function responses login a user and uses the `CreateToken` and `CreateAuth` function.<br>
`CreateToken` creates JWT with user claims. <br>
`CreateAuth` saves JWT-metadata in Redis.<br>
`Logout` logs out a user.<br>
`Refresh` refreshes the token.


