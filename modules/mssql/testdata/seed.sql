CREATE SCHEMA pizza_palace;
GO

CREATE TABLE pizza_palace.pizzas (
    ID INT PRIMARY KEY IDENTITY,
    ToppingName NVARCHAR(100),
    Deliciousness NVARCHAR(100) UNIQUE
);
GO

INSERT INTO pizza_palace.pizzas (ToppingName, Deliciousness) VALUES
    ('Pineapple', 'Controversial but tasty'),
    ('Pepperoni', 'Classic never fails')
GO
