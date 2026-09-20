<script setup lang="ts">
</script>

<template>
    <div class="Steps" v-reveal>
        <div class="Head">
            <h2>En 3 pasos estás jugando.</h2>
        </div>
        <div class="Cards sl-stagger">
            <div class="Card">
                <span class="Num">1</span>
                <div class="CardTxt">
                    <b>Descargá</b>
                    <small>Bajá el archivo de tu sistema desde el release de arriba.</small>
                </div>
            </div>
            <div class="Card">
                <span class="Num">2</span>
                <div class="CardTxt">
                    <b>Ejecutá</b>
                    <small>Abrí StepLauncher y dejá que detecte tu Java solo.</small>
                </div>
            </div>
            <div class="Card">
                <span class="Num">3</span>
                <div class="CardTxt">
                    <b>Creá tu cuenta y jugá</b>
                    <small>Una cuenta local u online, y a pulsar Jugar.</small>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped lang="scss">
.Steps{
    width: calc(100% - 10rem);
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 0 5rem 4rem 5rem;
    margin: auto;
    gap: 1.5rem;
    div{
        display:flex;
        justify-content:center;
        align-items:center;
    }
    .Head{
        h2{
            margin: 0;
            font-family: 'Lexend';
            font-size: 1.6rem;
            font-weight: 600;
        }
    }
    .Cards{
        gap: .8rem;
        align-items:stretch;
        width: 100%;
        max-width: 60rem;
        .Card{
            flex: 1;
            align-items:flex-start;
            gap: .7rem;
            padding: .9rem 1rem;
            border-radius: .7rem;
            border: 1px solid #ffffff14;
            background: #ffffff08;
            .Num{
                display:flex;
                justify-content:center;
                align-items:center;
                width: 2.2rem;
                height: 2.2rem;
                border-radius: .55rem;
                background: #fff;
                color: #000;
                font-family: 'Lexend';
                font-weight: 700;
                flex-shrink: 0;
            }
            .CardTxt{
                flex-direction:column;
                align-items:flex-start;
                text-align:left;
                gap: .3rem;
                b{
                    font-size: .82rem;
                    font-family: 'Lexend';
                }
                small{
                    font-size: .72rem;
                    opacity: .55;
                    line-height: 1.5;
                }
            }
        }
    }
}
@media (max-width: 1000px){
    .Steps{
        width: calc(100% - 3rem);
        padding: 0 1.5rem 3rem 1.5rem;
        .Cards{
            flex-direction:column;
        }
    }
}
</style>
