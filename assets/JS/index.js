 /*=============== UPDATING Like / Dislike ===============*/
 function like() {
 let cardsBody = document.querySelectorAll(".album-actions");
 let cartIcon = document.querySelector(".album-action span");
 cartIcon.classList.add("iconClass");
 let count = 1;
 cardsBody.forEach((cardBody) => {
   let btnSubmit = cardBody.children[1];
   btnSubmit.addEventListener("click", () => {
     cartIcon.innerHTML = count;
     count++;
   });
 });
 }
 