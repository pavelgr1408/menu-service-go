INSERT INTO menu_categories(id, name, display_order) VALUES
('burgers', 'Бургеры', 10),
('chicken', 'Курица', 20),
('potatoes-and-snacks', 'Картофель и закуски', 30),
('drinks', 'Напитки', 40),
('desserts', 'Десерты', 50),
('sauces', 'Соусы', 60);

INSERT INTO menu_items(
    id,
    category_id,
    name,
    short_description,
    full_description,
    image_url,
    price_amount,
    price_currency,
    price_formatted,
    status,
    calories_kcal,
    proteins_g,
    fats_g,
    carbohydrates_g,
    display_order
) VALUES
('big-burger', 'burgers', 'Большой бургер', 'Бургер с говяжьей котлетой, сыром, салатом и фирменным соусом', 'Большой бургер с говяжьей котлетой, сыром, свежим салатом, маринованными огурцами и фирменным соусом в булочке с кунжутом.', 'https://cdn.example.com/menu/items/big-burger.png', 29900, 'RUB', '299 ₽', 'AVAILABLE', 540, 27.50, 29.10, 42.40, 10),
('cheese-burger', 'burgers', 'Чизбургер', 'Бургер с сыром, говяжьей котлетой, луком и маринованными огурцами', 'Классический чизбургер с говяжьей котлетой, сыром, луком, маринованными огурцами, кетчупом и горчицей.', 'https://cdn.example.com/menu/items/cheese-burger.png', 15900, 'RUB', '159 ₽', 'AVAILABLE', 320, 16.30, 14.80, 31.20, 20),
('double-burger', 'burgers', 'Двойной бургер', 'Бургер с двумя говяжьими котлетами и сыром', 'Сытный бургер с двумя говяжьими котлетами, двумя ломтиками сыра, соусом и свежими овощами.', 'https://cdn.example.com/menu/items/double-burger.png', 34900, 'RUB', '349 ₽', 'UNAVAILABLE', 710, 39.70, 42.30, 44.80, 30),
('chicken-burger', 'chicken', 'Чикенбургер', 'Бургер с куриной котлетой, салатом и сливочным соусом', 'Чикенбургер с хрустящей куриной котлетой, листовым салатом, свежими овощами и сливочным соусом.', 'https://cdn.example.com/menu/items/chicken-burger.png', 21900, 'RUB', '219 ₽', 'AVAILABLE', 430, 22.20, 18.60, 39.50, 10),
('nuggets', 'chicken', 'Наггетсы', 'Куриные наггетсы в хрустящей панировке', 'Сочные куриные наггетсы в хрустящей панировке. Хорошо сочетаются с сырным или барбекю-соусом.', 'https://cdn.example.com/menu/items/nuggets.png', 18900, 'RUB', '189 ₽', 'AVAILABLE', 360, 21.00, 20.40, 23.80, 20),
('french-fries', 'potatoes-and-snacks', 'Картофель фри', 'Золотистый картофель фри с солью', 'Классический золотистый картофель фри с хрустящей корочкой и мягкой серединкой.', 'https://cdn.example.com/menu/items/french-fries.png', 11900, 'RUB', '119 ₽', 'AVAILABLE', 310, 4.20, 15.60, 39.90, 10),
('cola', 'drinks', 'Кола', 'Газированный напиток со вкусом колы', 'Охлаждённый газированный напиток со вкусом колы.', 'https://cdn.example.com/menu/items/cola.png', 9900, 'RUB', '99 ₽', 'AVAILABLE', 140, 0.00, 0.00, 35.00, 10),
('orange-juice', 'drinks', 'Апельсиновый сок', 'Апельсиновый сок без газа', 'Освежающий апельсиновый сок без газа.', 'https://cdn.example.com/menu/items/orange-juice.png', 12900, 'RUB', '129 ₽', 'AVAILABLE', 120, 1.20, 0.20, 28.60, 20),
('ice-cream', 'desserts', 'Мороженое', 'Ванильное мороженое в стаканчике', 'Нежное ванильное мороженое в удобном стаканчике.', 'https://cdn.example.com/menu/items/ice-cream.png', 10900, 'RUB', '109 ₽', 'AVAILABLE', 210, 4.10, 9.80, 27.20, 10),
('apple-pie', 'desserts', 'Яблочный пирожок', 'Горячий пирожок с яблочной начинкой', 'Хрустящий горячий пирожок с ароматной яблочной начинкой.', 'https://cdn.example.com/menu/items/apple-pie.png', 9900, 'RUB', '99 ₽', 'AVAILABLE', 260, 3.20, 12.40, 34.50, 20),
('cheese-sauce', 'sauces', 'Сырный соус', 'Порционный соус с сырным вкусом', 'Нежный порционный соус с выраженным сырным вкусом.', 'https://cdn.example.com/menu/items/cheese-sauce.png', 4900, 'RUB', '49 ₽', 'AVAILABLE', 80, 1.10, 7.20, 2.70, 10),
('barbecue-sauce', 'sauces', 'Соус барбекю', 'Порционный соус барбекю', 'Пикантный порционный соус барбекю с копчёными нотами.', 'https://cdn.example.com/menu/items/barbecue-sauce.png', 4900, 'RUB', '49 ₽', 'AVAILABLE', 60, 0.50, 0.20, 13.40, 20);
