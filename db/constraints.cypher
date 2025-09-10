CREATE CONSTRAINT UserEmailUnique IF NOT EXISTS 
FOR (user:User) 
REQUIRE user.email IS UNIQUE;

CREATE CONSTRAINT unique_reserved_date 
FOR (e:ReservedDay)
REQUIRE date(e.day) IS UNIQUE;

CREATE CONSTRAINT unique_res_date 
FOR (e:ReservedDay)
REQUIRE date(e.day) IS UNIQUE;
