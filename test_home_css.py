import re

with open('frontend/src/components/PowerFlow.vue', 'r') as f:
    content = f.read()

# Let's check the size of the Home node compared to others.
# Others are w-28 h-28 (112px).
# Our Home node is w-[120px] h-[120px]. And border is 120-104 = 16px. So 8px on each side.
# Double standard border which is border-[4px] (4px).

# Make sure inner circle is perfectly centered.
# `<div class="flex flex-col items-center justify-center w-[104px] h-[104px] bg-white dark:bg-gray-800 rounded-full">`
# Added `absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2` to ensure it stays in middle.

old_inner = '<div class="flex flex-col items-center justify-center w-[104px] h-[104px] bg-white dark:bg-gray-800 rounded-full">'
new_inner = '<div class="absolute flex flex-col items-center justify-center w-[104px] h-[104px] bg-white dark:bg-gray-800 rounded-full top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2">'

content = content.replace(old_inner, new_inner)

with open('frontend/src/components/PowerFlow.vue', 'w') as f:
    f.write(content)
print("Updated inner circle layout")
