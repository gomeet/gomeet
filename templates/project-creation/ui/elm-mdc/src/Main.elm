module Main exposing (..)

import Html exposing (Html, text)
import Material
import Material.Button as Button
import Material.Fab as Fab
import Material.Options as Options exposing (styled, css)
import Material.Typography as Typography


type alias Model =
    { mdc : Material.Model Msg
    , counter : Int
    }


defaultModel : Model
defaultModel =
    { mdc = Material.defaultModel
    , counter = 0
    }


type Msg
    = Mdc (Material.Msg Msg)
    | Increment
    | Decrement


main : Program Never Model Msg
main =
    Html.program
        { init = init
        , subscriptions = subscriptions
        , update = update
        , view = view
        }


init : ( Model, Cmd Msg )
init =
    ( defaultModel, Material.init Mdc )


subscriptions : Model -> Sub Msg
subscriptions model =
    Material.subscriptions Mdc model


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        Mdc msg_ ->
            Material.update Mdc msg_ model

        Decrement ->
            ( { model | counter = model.counter - 1 }, Cmd.none )

        Increment ->
            ( { model | counter = model.counter + 1 }, Cmd.none )


view : Model -> Html Msg
view model =
    Html.div []
        [ viewBtnMinus model
        , viewCounter model
        , viewBtnPlus model
        ]


viewCounter : Model -> Html Msg
viewCounter model =
    styled Html.h2
        [ css "position" "relative"
        , css "top" "-20px"
        , css "display" "block"
        , css "width" "100px"
        , css "height" "40px"
        , css "line-height" "40px"
        , css "float" "left"
        , css "text-align" "center"
        , Typography.display1
        , Typography.adjustMargin
        ]
        [ text <| toString model.counter ]


viewBtnMinus : Model -> Html Msg
viewBtnMinus model =
    viewBtn model 1 Decrement "remove"


viewBtnPlus : Model -> Html Msg
viewBtnPlus model =
    viewBtn model 2 Increment "add"


viewBtn : Model -> Int -> Msg -> String -> Html Msg
viewBtn model mdcIdx msg icon =
    Fab.view Mdc
        [ mdcIdx ]
        model.mdc
        [ Fab.ripple
        , Fab.mini
        , Options.onClick msg
        , css "float" "left"
        ]
        icon


viewBtn_ : Model -> Int -> Msg -> String -> Html Msg
viewBtn_ model mdcIdx msg label =
    Button.view Mdc
        [ mdcIdx ]
        model.mdc
        [ Button.ripple
        , Options.onClick msg
        ]
        [ text label ]
